#
# Based on the Dockerfile for device-camera-go
# https://github.com/edgexfoundry/device-camera-go
#   Dockerfile
# Modified by Fujitsu Limited
#
# Copyright (c) 2020 IOTech Ltd
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

ARG BASE=opencv-gpu-cuda-arm64
FROM ${BASE} AS builder

ARG MAKE="make build"
ARG PKG_BASE="make git gcc libc-dev libsodium-dev libzmq3-dev"

LABEL Name=edgex-image-app

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022: Fujitsu Limited'

COPY 90proxy /etc/apt/apt.conf.d
RUN apt update && apt install -y ${PKG_BASE}

WORKDIR /image-app

COPY go.mod .
RUN go mod download all

COPY . .
RUN ${MAKE}

FROM ${BASE}

COPY --from=builder /etc/apt/apt.conf.d/90proxy /etc/apt/apt.conf.d/90proxy
RUN apt update
RUN apt install -y ca-certificates libzmq5 dumb-init
RUN rm /etc/apt/apt.conf.d/90proxy

WORKDIR /
COPY --from=builder /image-app/image-app /image-app
COPY --from=builder /image-app/res/ /res/

EXPOSE 59741

ENTRYPOINT ["/image-app"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
