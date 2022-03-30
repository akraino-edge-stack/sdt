# COPYRIGHT 2022 FUJITSU LIMITED
#
# SPDX-License-Identifier: Apache-2.0
.PHONY: docker docker-arm run

docker:
	docker build \
	    --network=host \
		-f Dockerfile \
		-t edgexfoundry/shim-security-bootstrapper-amd64:latest \
		.

docker-arm:
	docker build \
	    --network=host \
		-f Dockerfile-arm \
		-t edgexfoundry/shim-security-bootstrapper-arm64:latest \
		.

run:
	docker run \
		-v edgex-init:/tmp/edgex-init:z \
		edgexfoundry/shim-security-bootstrapper-amd64:latest
