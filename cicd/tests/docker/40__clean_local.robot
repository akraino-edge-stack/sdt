*** Settings ***
Documentation     Test cleaning local Docker images
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Clean local Docker images
    ${output}=                 Execute Command       docker image ls --all
    ${lc}=                     Get Line Count        ${output}
    Should Be True             ${lc} > 2
    ${rc}=                     Execute Playbook      clean_local_images.yml
    Should Be Equal            ${rc}                      ${0}
    ${output}=                 Execute Command       docker image ls --all
    Should Not Contain         ${output}             redis
    Should Not Contain         ${output}             edgex
    Should Not Contain         ${output}             kube
    Should Not Contain         ${output}             flannel
    Should Not Contain         ${output}             kuiper
    Should Not Contain         ${output}             coredns
    Should Not Contain         ${output}             etcd
    Should Not Contain         ${output}             pause
    Should Not Contain         ${output}             consul

