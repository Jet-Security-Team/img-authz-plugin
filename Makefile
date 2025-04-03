PLUGIN_NAME ?= jet-isc/img-authz-plugin
PLUGIN_TAG ?= $(shell arch)
BUILD_DIR = PLUGIN_${PLUGIN_TAG}
REGISTRIES :=
NOTARY :=
NOTARY_ROOT_CA :=


all: clean rootfs create push

build: clean rootfs

local_install: clean rootfs create enable

clean:
	@echo " - Removing the local build cache ./${BUILD_DIR}"
	@rm -rf ./${BUILD_DIR}

rootfs:
	@echo " - Building the rootfs Docker image"
	@docker build -t ${PLUGIN_NAME}:rootfs .
	@echo " - Create rootfs folder at ./${BUILD_DIR}/rootfs"
	@mkdir -p ./${BUILD_DIR}/rootfs
	@echo " - Initialize container from ${PLUGIN_NAME}:rootfs"
	@docker create --name rootfs ${PLUGIN_NAME}:rootfs true
	@echo " - Exporting container filesystem into ./${BUILD_DIR}/rootfs"
	@docker export rootfs | tar -x -C ./${BUILD_DIR}/rootfs
	@echo " - Copying config.json to ./${BUILD_DIR}"
	@cp config.json ./${BUILD_DIR}/
	@echo " - Deleting build container rootfs"
	@docker rm -vf rootfs

create:
	@echo " - Removing existing plugin ${PLUGIN_NAME}:${PLUGIN_TAG} if exists"
	@docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	@echo " - Creating new plugin ${PLUGIN_NAME}:${PLUGIN_TAG} from ./${BUILD_DIR}"
	@docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./${BUILD_DIR}

push:
	@echo " - Publishing plugin ${PLUGIN_NAME}:${PLUGIN_TAG} (make sure access to chosen registry is provided)"
	@docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
