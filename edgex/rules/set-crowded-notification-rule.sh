#!/bin/bash

pod_ip=$(kubectl get ep -A | grep edgex | awk -F "[ :,]+" '{print $3}')
if [ -z $pod_ip ]; then
    echo "get pod_ip NG"
    exit 1
fi

array=(${pod_ip//\n/ })
if [ -z $array ]; then
    echo "cut pod_ip NG"
    exit 2
fi

stream_ok="Stream edgexAll is created."
rule_ok="Rule rule1 was created successfully."
ng_flag=0
for var in ${array[@]}
do
    #set stream
    stream=$(curl --noproxy $var -X "POST" "http://$var:59720/streams" -H "Content-Type: application/json" -d '{"sql": "create stream edgexAll() WITH (FORMAT=\"JSON\", TYPE=\"edgex\")"}')
    if [[ $stream != *$stream_ok* ]]; then
        echo "set stream NG. pod ip is $var"
        ng_flag=1
        continue
    fi

    #get stream
    curl --noproxy $var -X "GET" "http://$var:59720/streams/edgexAll"

    #set rule
    rule=$(curl --noproxy $var -X "POST" "http://$var:59720/rules" -H "Content-Type: application/json" \
  -d "{
  \"id\": \"rule1\",
  \"sql\": \"SELECT Crowded FROM edgexAll WHERE Crowded = true\",
  \"actions\": [
    {
      \"rest\": {
        \"url\": \"http://$var:59860/api/v2/notification\",
        \"method\": \"post\",
        \"dataTemplate\": \"[{\\\"apiVersion\\\": \\\"v2\\\", \\\"notification\\\": { \\\"category\\\": \\\"daily-notify\\\", \\\"content\\\": \\\"crowding detected\\\", \\\"contentType\\\": \\\"string\\\", \\\"sender\\\": \\\"edgex-kuiper\\\", \\\"severity\\\": \\\"NORMAL\\\" } }]\",
        \"sendSingle\": false
      }
    },
    {
      \"log\":{}
    }
  ]
}")
    if [[ $rule != *$rule_ok* ]]; then
        echo "set rule NG. pod ip is $var"
        ng_flag=1
        continue
    fi

    #get rule
    curl --noproxy $var -X "GET" "http://$var:59720/rules/rule1"
done

if [ $ng_flag -ne 0 ]; then
    exit 3
else
    exit 0
fi