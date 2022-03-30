# COPYRIGHT 2022 FUJITSU LIMITED
#
# SPDX-License-Identifier: Apache-2.0
.PHONY: docker docker-arm

docker:
	docker build \
		-f Dockerfile \
		-t kong-patched-amd64:latest \
		.

docker-arm:
	docker build \
		-f Dockerfile-arm \
		-t kong-patched-arm64:latest \
		.
