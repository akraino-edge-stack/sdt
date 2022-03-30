*** Settings ***
Documentation     Test setup of ARM build node
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Setup Build Node
    ${rc}=             Execute CICD Playbook
    ...                setup_arm_build.yml
    Should Be Equal    ${rc}                ${0}
    # TODO: Log in to ARM build server to confirm status
