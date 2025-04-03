#!/bin/sh

run_container() {
  # $1 is the image name to be executed
  timeout 30s docker run -d $1 echo run_container successful
}

pull_image() {
  # $1 is the image to be pulled
  timeout 30s docker pull -q $1
}

AUTHZ_IMG="alpine"
AUTHZ_IMG_TAG="alpine:latest"
AUTHZ_IMG_FULL="docker.io/library/alpine:latest"
UNAUTHZ_IMG="my.unauthz.docker.registry/alpine:latest"


test_docker_pull_is_allowed() {
  assertTrue "Failed to PULL authorized ${AUTHZ_IMG_TAG}" "pull_image ${AUTHZ_IMG_TAG}"
}

test_docker_run_is_allowed() {
  assertTrue "Failed to RUN authorized ${AUTHZ_IMG_TAG}" "run_container ${AUTHZ_IMG_TAG}"
}

test_pull_is_not_allowed_when_registry_is_not_authorized(){
  assertFalse "Failed to BLOCK PULL of unauthorized ${UNAUTHZ_IMG}" "pull_image ${UNAUTHZ_IMG}"
}

test_run_is_not_allowed_when_registry_is_not_authorized(){
  assertFalse "Failed to BLOCK RUN of unauthorized ${UNAUTHZ_IMG}" "run_container ${UNAUTHZ_IMG}"
}

test_pull_is_allowed_when_registry_is_authorized(){
  assertTrue "Failed to PULL authorized with full path ${AUTHZ_IMG_FULL}" "pull_image ${AUTHZ_IMG_FULL}"
}

test_run_is_allowed_when_registry_is_authorized(){
  assertTrue "Failed to RUN authorized with full path ${AUTHZ_IMG_FULL}" "run_container ${AUTHZ_IMG_FULL}"
}

test_pull_is_allowed_when_tag_not_specified(){
  assertTrue "Failed to PULL authorized image without tag ${AUTHZ_IMG}" "pull_image ${AUTHZ_IMG}"
}

test_run_is_allowed_when_tag_not_specified(){
  assertTrue "Failed to RUN authorized image without tag ${AUTHZ_IMG}" "run_container ${AUTHZ_IMG}"
}
