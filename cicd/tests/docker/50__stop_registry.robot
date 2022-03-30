*** Settings ***
Documentation     Test stopping the private Docker registry
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Stop Docker registry
    ${output}=                 Execute Command       docker ps
    Should Match Regexp        ${output}             registry
    ${rc}=                     Execute Playbook      stop_registry.yml
    Should Be Equal            ${rc}                 ${0}
    ${output}=                 Execute Command       docker ps
    Should Not Match Regexp    ${output}             registry

