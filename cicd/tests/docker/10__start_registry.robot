*** Settings ***
Documentation     Test starting the private Docker registry
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Start Docker registry
    ${output}=            Execute Command       docker ps
    Should Not Contain    ${output}             registry
    ${rc}=                Execute Playbook      start_registry.yml
    Should Be Equal       ${rc}                 ${0}
    ${output}=            Execute Command       docker ps
    Should Contain        ${output}             registry

