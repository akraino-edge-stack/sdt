#!/bin/sh
#
# Delete the gateway API user created using gw-user.sh
#
# Usage: del-gw-user.sh <pod>
#
# COPYRIGHT 2022 FUJITSU LIMITED
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
if [ $# != 1 ]; then
    echo "Usage: $0 <pod>"
    exit
fi

USER=gwuser
POD=$1

KONGJWT=`kubectl exec $POD -c edgex-security-proxy-setup -- cat /tmp/edgex/secrets/security-proxy-setup/kong-admin-jwt`

kubectl exec $POD -c edgex-security-proxy-setup -- /edgex/secrets-config proxy deluser --user $USER --jwt $KONGJWT
