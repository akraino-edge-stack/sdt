# COPYRIGHT 2022 FUJITSU LIMITED
#
# SPDX-License-Identifier: Apache-2.0
FROM edgexfoundry/security-secretstore-setup:2.1.0 AS original
FROM alpine:3.14
LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

RUN apk add --update --no-cache ca-certificates dumb-init su-exec

WORKDIR /

COPY --from=original /res-file-token-provider/ /res-file-token-provider
COPY --from=original /res/ /res

COPY --from=original /security-file-token-provider .
COPY --from=original /security-secretstore-setup .
COPY --from=original /vault/config/ /vault/config

COPY --from=original /usr/local/bin/entrypoint.sh /usr/local/bin/
RUN ln -s /usr/local/bin/entrypoint.sh /

# Workaround because sharing /vault/config hides the contents
# There's just a directory "assets" there.
RUN sed -i -e '2 i mkdir -p /vault/config/assets && chown -Rh 100:1000 /vault/' /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
