*** Settings ***
Documentation     Test adding edge nodes to the cluster via Ansible
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Add edge nodes to cluster
    ${output}=                 Kubectl             get nodes
    Should Match Regexp        ${output}           ^NAME
    ${lc}=                     Get Line Count      ${output}
    Should Be Equal            ${lc}               ${2}                 msg='Only the controller node should be configured'
    ${rc}=                     Execute Playbook    join_cluster.yml
    Should Be Equal            ${rc}               ${0}
    Wait Until All Nodes Ready
    ${output}=                 Kubectl             get nodes
    ${lc}=                     Get Line Count      ${output}
    Should Be True             ${lc} > 2

