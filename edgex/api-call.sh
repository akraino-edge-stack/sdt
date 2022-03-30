#!/bin/sh
#
# Call a specified gateway through the gateway server
#
# Usage: api-call.sh <address> <jwt> <path> ...
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
if [ $# -lt 3 ]; then
    echo "Usage: $0 <address> <jwt> <path> [additional curl options ...]"
    exit
fi

ADDR=$1
JWT=$2
APIPATH=$3

shift 3

curl --noproxy $ADDR -k -H "Authorization: Bearer $JWT" $* https://$ADDR:8443/$APIPATH
