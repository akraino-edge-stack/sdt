#!/bin/sh
#
# Create a gateway API user and JWT with the specified key.
# EdgeX services need to be up and running to do this.
# The user needs to be created every time the services are restarted.
#
# Usage: gw-user.sh <node> <pod>
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
if [ $# != 2 ]; then
    echo "Usage: $0 <node> <pod>"
    exit
fi

USER=gwuser
KEY=$USER.pub
PKEY=$USER.key
NODE=$1
POD=$2

KONGJWT=`kubectl exec $POD -c edgex-security-proxy-setup -- cat /tmp/edgex/secrets/security-proxy-setup/kong-admin-jwt`

openssl ecparam -name prime256v1 -genkey -noout -out $PKEY
openssl ec -out $KEY < $PKEY
chmod a+r $KEY

# NOTE: The key file used to log into the edge node is hard coded here
scp -i ~/.ssh/edge $KEY edge@$NODE:/opt/lfedge/volumes/edgex/secrets/security-proxy-setup/$KEY

ID=`uuidgen`
echo "User ID is $ID"
echo

kubectl exec $POD -c edgex-security-proxy-setup -- /edgex/secrets-config proxy adduser --token-type jwt --id "$ID" --algorithm ES256 --public_key /tmp/edgex/secrets/security-proxy-setup/$KEY --user $USER --jwt "$KONGJWT"

echo
JWT=`docker run --rm -v $PWD:/host:ro -u "$UID" --entrypoint "/edgex/secrets-config" master:5000/edgexfoundry/security-proxy-setup:2.1.0 -- proxy jwt --id "$ID" --algorithm ES256 --private_key /host/$PKEY`

echo "JWT: $JWT"

rm $KEY $PKEY
echo "Done"
