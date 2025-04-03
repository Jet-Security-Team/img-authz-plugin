// Docker Image Authorization Plugin.
// Allows docker images to be fetched from a list of authorized registries only.
// AUTHOR: Chaitanya Prakash N <cpdevws@gmail.com>
package main

import (
	"flag"
	"fmt"
	"log"
  	"net/url"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/docker/go-plugins-helpers/authorization"
)

const (
	defaultDockerHost = "unix:///var/run/docker.sock"
	pluginSocket      = "/run/docker/plugins/img-authz-plugin.sock"
)

var (
	flDockerHost         = flag.String("host", defaultDockerHost, "Specifies the host where docker daemon is running")
	authorizedRegistries stringslice
	Version              string
	Build                string
)

func main() {

	log.Println("Plugin Version:", Version, "Build: ", Build)

	// Fetch the registry from env
	var defaultRegistry = "docker.io"
	authorizedRegistry, _ := os.LookupEnv("REGISTRY")

	if len(authorizedRegistry) == 0 {
		authorizedRegistry = defaultRegistry
		log.Println("REGISTRY was not set. Defaulting to:", defaultRegistry)
	}

	// Fetch the notary from env
	var defaultNotary = "https://notary.docker.io"
	authorizedNotary, _ := os.LookupEnv("NOTARY")

	if len(authorizedNotary) == 0 {
		authorizedNotary = defaultNotary
		log.Println("NOTARY Server was not set. Defaulting to:", defaultNotary)
	}

	if !strings.HasPrefix(authorizedNotary, "https://") {
		authorizedNotary = "https://" + authorizedNotary
	}

	// Fetch the notary RootCA from env
	notaryRootCA, _ := os.LookupEnv("NOTARY_ROOT_CA")
	var notaryRootCAFile string
	if len(notaryRootCA) == 0 {
		notaryRootCAFile = ""

		log.Println("NOTARY_ROOT_CA was not passed. Assuming the Notary server has been signed by a recognized public CA!")
	} else{
		notaryURL, _ := url.ParseRequestURI(authorizedNotary)

		var notaryRootCAFolder = fmt.Sprintf("/root/.docker/tls/%s", notaryURL.Host)
		notaryRootCAFile = fmt.Sprintf("%s/root-ca.crt", notaryRootCAFolder)
		os.MkdirAll(notaryRootCAFolder, os.ModePerm)

		f, err := os.Create(notaryRootCAFile)
		errt := f.Truncate(0)
		if err != nil || errt != nil {
			log.Fatal(err, errt)
		}

		defer f.Close()
		_, err2 := f.WriteString(notaryRootCA)
		if err2 != nil {
			log.Fatal(err2)
		}

		log.Println("Notary Root CA: ", notaryRootCAFile)
	}

	// Convert authorized registries into a map for efficient lookup
	// NB! Although, only single registry is expected at the moment,
	//     wee keep registries map for the later extensibility.
	registries := make(map[string]bool)
	log.Println("Authorized registry:", authorizedRegistry)
	registries[authorizedRegistry] = true

	log.Println("Authorized notary: ", authorizedNotary)

	// Create image authorization plugin
	plugin, err := newPlugin(*flDockerHost, registries, authorizedNotary, notaryRootCAFile)
	if err != nil {
		log.Fatal(err)
	}

	// Start service handler on the local sock
	u, _ := user.Lookup("root")
	gid, _ := strconv.Atoi(u.Gid)
	handler := authorization.NewHandler(plugin)
	if err := handler.ServeUnix(pluginSocket, gid); err != nil {
		log.Fatal(err)
	}
}
