#
# Based on the example application service makefile
# https://github.com/edgexfoundry/app-functions-sdk-go/
#   app-service-template/Makefile
# Modified by Fujitsu Limited
#
# Copyright (c) 2021 Intel Corporation
# Copyright (c) 2022 Fujitsu Limited
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.PHONY: build clean docker opencv tidy

GO=CGO_ENABLED=1 go

APPVERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)

# This pulls the version of the SDK from the go.mod file. If the SDK is the only required module,
# it must first remove the word 'required' so the offset of $2 is the same if there are multiple required modules
# SDKVERSION=$(shell cat ./go.mod | grep 'github.com/edgexfoundry/app-functions-sdk-go/v2 v' | sed 's/require//g' | awk '{print $$2}')

MICROSERVICE=image-app
IMAGE=edgexfoundry/image-app
# GOFLAGS=-ldflags "-X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.SDKVersion=$(SDKVERSION) -X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.ApplicationVersion=$(APPVERSION)"

build: tidy
	$(GO) build $(GOFLAGS) -o $(MICROSERVICE)

tidy:
	go mod tidy

APT_PROXY_FILE=90proxy

$(APT_PROXY_FILE):
	if [ "${HTTP_PROXY}" = "" ]; then touch $@ ; else echo Acquire::http::Proxy \"${HTTP_PROXY}\"\; > $@ ; fi

SET_PROXY_FILE=set_proxy

$(SET_PROXY_FILE):
	if [ "${HTTPS_PROXY}" = "" ]; then touch $@ ; else echo export https_proxy=${HTTPS_PROXY} > $@ ; fi

docker: $(APT_PROXY_FILE)
	docker build \
		--network=host \
		--build-arg HTTP_PROXY=$(HTTP_PROXY) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		-f Dockerfile \
		-t $(IMAGE)-arm64 \
		.

opencv: $(APT_PROXY_FILE) $(SET_PROXY_FILE)
	docker build \
		--network=host \
		--build-arg HTTP_PROXY=$(HTTP_PROXY) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		-f Dockerfile.opencv-gpu-cuda \
		-t opencv-gpu-cuda-arm64 \
		.

clean:
	rm -f $(MICROSERVICE)
	rm -f $(APT_PROXY_FILE)
