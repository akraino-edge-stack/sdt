*** Settings ***
Documentation     Test building and pushing images
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Build Images
    ${rc}=             Execute CICD Playbook    build_images.yml
    Should Be Equal    ${rc}                    ${0}

Push Images To Registry
    ${rc}=             Execute CICD Playbook    push_images.yml
    Should Be Equal    ${rc}                    ${0}
