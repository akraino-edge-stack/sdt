*** Settings ***
Documentation     Test removing the private Docker registry
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Remove Docker registry
    ${output}=                 Execute Command            ssh master docker ps
    Should Not Contain         ${output}                  registry
    ${output}=                 Execute Command            ssh master docker container ls --all
    Should Contain             ${output}                  registry
    ${rc}=                     Execute Playbook           remove_registry.yml
    Should Be Equal            ${rc}                      ${0}
    ${output}=                 Execute Command            ssh master docker container ls --all
    Should Not Contain         ${output}                  registry

