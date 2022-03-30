*** Settings ***
Documentation     Test pulling upstream Docker images
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Pull upstream Docker images
    ${output}=             Execute Command       docker ps
    Should Match Regexp    ${output}             registry
    ${rc}=                 Execute Playbook      pull_upstream_images.yml
    Should Be Equal        ${rc}                 ${0}
    ${output}=             Execute Command       docker image ls --all
    Should Contain         ${output}             redis
    Should Contain         ${output}             edgex
    Should Contain         ${output}             kube
    Should Contain         ${output}             flannel
    Should Contain         ${output}             kuiper
    Should Contain         ${output}             coredns
    Should Contain         ${output}             etcd
    Should Contain         ${output}             pause
    Should Contain         ${output}             consul
    ${output}=             Execute Command       curl --noproxy "*" http://127.0.0.1:5000/v2/_catalog
    Should Contain         ${output}             redis
    Should Contain         ${output}             edgex
    Should Contain         ${output}             kube
    Should Contain         ${output}             flannel
    Should Contain         ${output}             kuiper
    Should Contain         ${output}             coredns
    Should Contain         ${output}             etcd
    Should Contain         ${output}             pause
    Should Contain         ${output}             consul
