.PHONY: build test clean docker unittest lint

ARCH=$(shell uname -m)
GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=cmd/device-lora
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
DOCKER_TAG=$(VERSION)-dev

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-lora.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

tidy:
	go mod tidy

build: $(MICROSERVICES)

cmd/device-lora:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd


# NOTE: Configure the location of an apt proxy configuration file if necessary.
# If the file does not exist no proxy will be used.
APT_PROXY_FILE=90proxy
APT_PROXY_PATH=/etc/apt/apt.conf.d/$(APT_PROXY_FILE)

$(APT_PROXY_FILE):
	if [ -f $(APT_PROXY_PATH) ]; then cp $(APT_PROXY_PATH) ./$(APT_PROXY_FILE) ; else touch ./$(APT_PROXY_FILE); fi

docker: $(APT_PROXY_FILE)
	docker build \
		--network=host \
		--build-arg HTTP_PROXY=$(HTTP_PROXY) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		-f Dockerfile \
		-t edgexfoundry/device-lora-amd64:latest \
		.

docker-arm: $(APT_PROXY_FILE)
	docker build \
		--network=host \
		--build-arg HTTP_PROXY=$(HTTP_PROXY) \
		--build-arg HTTPS_PROXY=$(HTTPS_PROXY) \
		-f Dockerfile-arm \
		-t edgexfoundry/device-lora-arm64:latest \
		.

clean:
	rm -f $(MICROSERVICES)
	rm -f $(APT_PROXY_FILE)
