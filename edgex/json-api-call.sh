#!/bin/sh
#
# Send JSON to an API through the gateway
# NOTE: Use -X PUT or -X POST as appropriate in the additional options
#
# Usage: json-api-call.sh <address> <jwt> <path> <json> ...
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
if [ $# -lt 4 ]; then
    echo "Usage: $0 <address> <jwt> <path> <json> [curl options...]"
    exit
fi

ADDR=$1
JWT=$2
APIPATH=$3
JSON=$4
JSONFILE=tmp.json

shift 4

echo "$JSON" > $JSONFILE
curl --noproxy $ADDR -k -H "Authorization: Bearer $JWT" -H "Content-type: application/json" $* -d @$JSONFILE https://$ADDR:8443/$APIPATH
rm $JSONFILE
