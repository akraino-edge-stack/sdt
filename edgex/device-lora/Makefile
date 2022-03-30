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

docker:
	docker build \
	    --network=host \
		-f Dockerfile \
		-t edgexfoundry/device-lora-amd64:latest \
		.

docker-arm:
	docker build \
	    --network=host \
		-f Dockerfile-arm \
		-t edgexfoundry/device-lora-arm64:latest \
		.

clean:
	rm -f $(MICROSERVICES)
