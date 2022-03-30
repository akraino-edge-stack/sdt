*** Settings ***
Documentation     Test setup of main build node
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Setup Build Node
    ${rc}=             Execute CICD Become Playbook
    ...                setup_build.yml
    Should Be Equal    ${rc}                ${0}
    ${output}=         Execute Command      apt list --installed
    Should Contain     ${output}            docker.io
    Should Contain     ${output}            make
    Directory Should Exist
    ...                edgexfoundry/app-functions-sdk-go
    Directory Should Exist
    ...                edgexfoundry/device-sdk-go
    File Should Exist
    ...                /usr/local/go/bin/go
