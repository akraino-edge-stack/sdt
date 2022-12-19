#
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

# NOTE: Using Debian bullseye image instead of alpine because of difficulties
# with apk and proxy settings
# ARG BASE=golang:1.16-alpine3.14
ARG BASE=golang:1.16-bullseye
FROM ${BASE} AS builder

ARG MAKE=make build

WORKDIR /device-lora

# RUN apk add --update --no-cache make git gcc libc-dev zeromq-dev libsodium-dev
COPY 90proxy /etc/apt/apt.conf.d
RUN apt update
RUN apt install -y make git gcc libc-dev libsodium-dev libzmq3-dev

COPY go.mod .
RUN go mod download all
COPY . .

RUN make cmd/device-lora

# Next image - Copy built Go binary into new workspace
# FROM alpine:3.14
FROM debian:bullseye
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022: Fujitsu Limited'

# RUN apk add --update --no-cache zeromq dumb-init
COPY --from=builder /etc/apt/apt.conf.d/90proxy /etc/apt/apt.conf.d/90proxy
RUN apt update
RUN apt install -y ca-certificates libzmq5 dumb-init
RUN rm /etc/apt/apt.conf.d/90proxy

WORKDIR /
COPY --from=builder /device-lora/cmd/ /

EXPOSE 59999

ENTRYPOINT ["/device-lora"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
