// Docker Image Authorization Plugin.
// Allows docker images to be fetched from a list of authorized registries only.
// AUTHOR: Chaitanya Prakash N <cpdevws@gmail.com>
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	dockerapi "github.com/docker/docker/api"
	dockercontainer "github.com/docker/docker/api/types/container"
  dockerswarm "github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/authorization"
)

// ImgAuthZPlugin Image Authorization Plugin struct definition
type ImgAuthZPlugin struct {
	// Docker client
	client *dockerclient.Client
	// Map of authorized registries
	authorizedRegistries map[string]bool
	// Number of authorized registries
	numAuthorizedRegistries int
	// List of authorized registries as string
	authRegistriesAsString string
	// Authorized notary
	authorizedNotary string
	// File holding the Root CA for the notary server
	authorizedNotaryRootCAFile string
}

// Returns the list of authorized registries as string
func authRegistries(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// Create a new image authorization plugin
func newPlugin(dockerHost string, registries map[string]bool, notary string, notaryRootCAFile string) (*ImgAuthZPlugin, error) {
	client, err := dockerclient.NewClient(dockerHost, dockerapi.DefaultVersion, nil, nil)

	if err != nil {
		return nil, err
	}

	return &ImgAuthZPlugin{
		client:                  client,
		authorizedRegistries:    registries,
		authorizedNotary:        notary,
		authorizedNotaryRootCAFile:	notaryRootCAFile,
		numAuthorizedRegistries: len(registries),
		authRegistriesAsString:  authRegistries(registries)}, nil
}

// Returns true if there are any authorized registries configured.
// Otherwise, returns false
func (plugin *ImgAuthZPlugin) hasAuthorizedRegistries() bool {
	return (plugin.numAuthorizedRegistries > 0)
}

// Parses the docker client command to determine the requested registry used in the command.
// If a registry is used in the command (i.e. docker pull or docker run commands), then the registry url and true is returned.
// Otherwise, returns empty string and false.
func (plugin *ImgAuthZPlugin) getRequestedRegistry(req authorization.Request, reqURL *url.URL) (string, string, bool) {

	image := ""
	registry := ""
	cleanSetRegistry := strings.TrimRight(plugin.authRegistriesAsString, "/")
	defaultEmptyRegistry := "docker.io"

	// docker run
	if strings.HasSuffix(reqURL.Path, "/containers/create") {
		var config dockercontainer.Config
		json.Unmarshal(req.RequestBody, &config)
		image = config.Image

    log.Println("Analysing container creation request for image: ", image)
	}

	// docker pull
	if strings.HasSuffix(reqURL.Path, "/images/create") {
		image = reqURL.Query().Get("fromImage")
		tag := reqURL.Query().Get("tag")
		if len(tag) > 0 {
			image = image + ":" + tag
		}

		log.Println("Analysing image pull for: ", image)
	}

  // docker service
	if strings.HasSuffix(reqURL.Path, "/services/create") {
    var config dockerswarm.ServiceSpec
    json.Unmarshal(req.RequestBody, &config)
    image = strings.Split(config.TaskTemplate.ContainerSpec.Image, "@")[0]

    log.Println("Analysing service creation with image: ", image)
  }

	if len(image) > 0 {
		// If no registry is specfied, assume it is the dockerhub!

		// because Docker removes docker.io form the fromImage query,
		// we cannot know whether the user has passed the registry in the
		// image name or not.
		//
		// Examples:
		//  - docker.io/library/ubuntu is valid (fromImage = "library/ubuntu")
		//  - docker.io/ubuntu is valid (fromImage = "ubuntu")
		//  - myregistry:5000/img is valid (fromImage = myregistry:5000/img)
		//  - name1/name2/name3/name4 is also valid...and we have no idea whether name is a registry, repo name, sub-repo name, etc.

		// So let's instead check of the existence of the registry in fromImage
		if strings.HasPrefix(image, cleanSetRegistry + "/") {
			// then we are good to go
			return image, plugin.authRegistriesAsString, true
		} else{
			// in this case, we might be dealing with docker.io,
			// but Docker has removed the registry from fromImage
			if cleanSetRegistry == defaultEmptyRegistry {
				// there's no way to know if the prefix is a registry hostname
				// or an image name

				// So, let's prepend the image with docker.io, and if fromImage
				// does indeed container a private registry, the Notary check
				// will then fail
				return cleanSetRegistry + "/" + image,
						plugin.authRegistriesAsString, true
			} else {
				// then we assume the requested registry is the prefix
				// even if it isn't, we are not using docker.io so we're good
				idx := strings.Index(image, "/")
				if idx != -1 {
					return image, image[0:idx], true
				} else {
					// fromImage is a single string with no "/"
					// just assume it is docker.io
					return image, defaultEmptyRegistry, true
				}
			}
		}
	}

	return image, registry, false
}

// AuthZReq Authorizes the docker client command.
// Non registry related commands are allowed by default.
// If the command uses a registry, the command is allowed only if the registry is authorized.
// Otherwise, the request is denied!
func (plugin *ImgAuthZPlugin) AuthZReq(req authorization.Request) authorization.Response {
	// Parse request and the request body
	reqURI, _ := url.QueryUnescape(req.RequestURI)
	reqURL, _ := url.ParseRequestURI(reqURI)

	// Find out the requested registry and whether or not a registry is present in the client command
	requestedImage, requestedRegistry, isRegistryCommand := plugin.getRequestedRegistry(req, reqURL)

	// Docker command do not involve registries
	if isRegistryCommand == false {
		// Allowed by default!
		log.Println("[ALLOWED] Not a registry command:", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: true}
	}

	// There are no authorized registries.
	if plugin.hasAuthorizedRegistries() == false {
		// So, deny the request by default!
		log.Println("[DENIED] No authorized registries", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "No authorized registries configured"}
	}

	// Verify that registry requested is authorized
	if plugin.authorizedRegistries[requestedRegistry] == true {
		// Is an authorized registry: Allow!
		log.Println("[ALLOWED] Registry:", requestedRegistry, req.RequestMethod, reqURL.String())
		// return authorization.Response{Allow: true}
	} else {
		// Oops.. The requested registry is not authorized. Deny the request!
		log.Println("[DENIED] Registry:", requestedRegistry, req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "You can only use docker images from the following authorized registries: " + plugin.authRegistriesAsString}
	}

	// There are no authorized notaries.
	if len(plugin.authorizedNotary) == 0 {
		// So, deny the request by default!
		log.Println("[DENIED] No authorized notaries configured", req.RequestMethod, reqURL.String())
		return authorization.Response{Allow: false, Msg: "No authorized notaries configured"}
	}

	// Enforce DCT
	log.Println("Enforcing DCT on", requestedImage, ". Request:", req.RequestMethod, reqURL.String())

	// we need to make sure the requested image is the the FQDN format,
	// and that for official images, the repo "library" is included
	trimmedImage := strings.TrimPrefix(strings.TrimPrefix(
		requestedImage,
		plugin.authRegistriesAsString), "/")

	// after trimming the registry, we need to consider two types of image references:
	//    - repo/image:tag, such as nuvla/api:latest, or
	//	  - image:tag, used for official images, who omit the "library" repo, like alpine:latest
	if (len(strings.Split(trimmedImage, "/")) == 1 && plugin.authRegistriesAsString == "docker.io") {
		// then the format is missing the repo "library"
		trimmedImage = "library/" + trimmedImage
	}

	imageTag := strings.Split(trimmedImage, ":")

	var tag string
	if len(imageTag) > 1 {
		tag = imageTag[1]
	} else {
		tag = "latest"
	}

	// reconstruct the GUN based on the string manipulations from above
	image := strings.TrimRight(plugin.authRegistriesAsString, "/") +
		"/" + imageTag[0]

	log.Println("Checking Notary server", plugin.authorizedNotary,
		"for trust on image:", image, tag)

	var cmd *exec.Cmd
	var executeThis string
	if len(plugin.authorizedNotaryRootCAFile) > 0 {
		executeThis = fmt.Sprintf("/go/bin/notary -s %s -d /root/.docker/trust --tlscacert %s lookup %s %s",
			plugin.authorizedNotary,
			plugin.authorizedNotaryRootCAFile,
			image,
			tag)
	} else {
		executeThis = fmt.Sprintf("/go/bin/notary -s %s -d /root/.docker/trust lookup %s %s",
			plugin.authorizedNotary,
			image,
			tag)
	}

	log.Println("Notary command: ", executeThis)
	cmd = exec.Command("sh", "-c", executeThis)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("[DENIED]", image+":"+tag, ". Reason:", string(out))
		log.Println(err)
		return authorization.Response{Allow: false, Msg: string(out)}
	}
	log.Println("[ALLOWED]", image+":"+tag)
	return authorization.Response{Allow: true}
}

// AuthZRes Authorizes the docker client response.
// All responses are allowed by default.
func (plugin *ImgAuthZPlugin) AuthZRes(req authorization.Request) authorization.Response {
	// Allowed by default.
	return authorization.Response{Allow: true}
}
