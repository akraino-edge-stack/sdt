*** Settings ***
Documentation     Test removing edge nodes from the cluster via Ansible
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Remove edge nodes to cluster
    ${output}=                 Execute Command     kubectl get nodes
    Should Match Regexp        ${output}           ^NAME
    ${lc}=                     Get Line Count      ${output}
    Should Be True             ${lc} > 2
    ${rc}=                     Execute Playbook    delete_from_cluster.yml
    Should Be Equal            ${rc}               ${0}
    Wait Until All Nodes Ready
    ${output}=                 Execute Command     kubectl get nodes
    ${lc}=                     Get Line Count      ${output}
    Should Be Equal            ${lc}               ${2}
