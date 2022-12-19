# COPYRIGHT 2022 FUJITSU LIMITED
#
# SPDX-License-Identifier: Apache-2.0
FROM edgexfoundry/security-secretstore-setup:2.1.0
LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

WORKDIR /

# Workaround because sharing /vault/config hides the contents
# There's just a directory "assets" there.
RUN sed -i -e '2 i mkdir -p /vault/config/assets && chown -Rh 100:1000 /vault/' /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
