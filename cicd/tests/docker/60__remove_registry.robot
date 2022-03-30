*** Settings ***
Documentation     Test removing the private Docker registry
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Remove Docker registry
    ${output}=                 Execute Command            docker ps
    Should Not Match Regexp    ${output}                  registry
    ${output}=                 Execute Command            docker container ls --all
    Should Match Regexp        ${output}                  registry
    ${rc}=                     Execute Playbook           remove_registry.yml
    Should Be Equal            ${rc}                      ${0}
    ${output}=                 Execute Command            docker container ls --all
    Should Not Match Regexp    ${output}                  registry

