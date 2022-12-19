*** Settings ***
Documentation     Test deleting EdgeX application via Ansible
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Execute Command
    ${rc}=                     Execute Playbook    edgex_stop.yml
    Should Be Equal            ${rc}               ${0}
Wait Until All Pods Deleting
    ${output}=                 Kubectl             get pods
    ${lc}=                     Get Line Count      ${output}
    Should Be Equal            ${lc}               ${0}              
