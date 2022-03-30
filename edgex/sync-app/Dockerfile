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
# Adapted to build sync-app by Fujitsu Limited

#build stage
ARG BASE=golang:1.16-alpine3.14
FROM ${BASE} AS builder

ARG ALPINE_PKG_BASE="make git gcc libc-dev libsodium-dev zeromq-dev"
ARG ALPINE_PKG_EXTRA=""

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}
WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

ARG MAKE="make build"
RUN $MAKE

#final stage
FROM alpine:3.14
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022: Fujitsu Limited'
LABEL Name=sync-app Version=${VERSION}

# dumb-init is required as security-bootstrapper uses it in the entrypoint script
RUN apk add --update --no-cache ca-certificates zeromq dumb-init

COPY --from=builder /app/res/ /res/
COPY --from=builder /app/sync-app /sync-app

EXPOSE 59740

ENTRYPOINT ["/sync-app"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
