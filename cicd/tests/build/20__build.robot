*** Settings ***
Documentation     Test building and pushing images
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Build x86 Images
    ${rc}=             Execute CICD Playbook    build_amd64.yml
    Should Be Equal    ${rc}                    ${0}

Build ARM Images
    ${rc}=             Execute CICD Playbook    build_arm64.yml
    Should Be Equal    ${rc}                    ${0}

Push Images To Registry
    ${rc}=             Execute CICD Playbook    push_images.yml
    Should Be Equal    ${rc}                    ${0}
