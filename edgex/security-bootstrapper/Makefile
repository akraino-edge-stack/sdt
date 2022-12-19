# COPYRIGHT 2022 FUJITSU LIMITED
#
# SPDX-License-Identifier: Apache-2.0
.PHONY: docker docker-arm run

docker:
	docker pull edgexfoundry/security-bootstrapper:2.1.0
	docker-copyedit.py FROM edgexfoundry/security-bootstrapper:2.1.0 INTO shim-security-bootstrapper-amd64:base REMOVE VOLUME /edgex-init
	docker build \
		--network=host \
		-f Dockerfile \
		-t edgexfoundry/shim-security-bootstrapper-amd64:latest \
		.

docker-arm:
	docker pull edgexfoundry/security-bootstrapper-arm64:2.1.0
	docker-copyedit.py FROM edgexfoundry/security-bootstrapper-arm64:2.1.0 INTO shim-security-bootstrapper-arm64:base REMOVE VOLUME /edgex-init
	docker build \
		--network=host \
		-f Dockerfile-arm \
		-t edgexfoundry/shim-security-bootstrapper-arm64:latest \
		.

run:
	docker run \
		-v edgex-init:/tmp/edgex-init:z \
		edgexfoundry/shim-security-bootstrapper-amd64:latest
