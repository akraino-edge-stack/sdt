*** Settings ***
Documentation     Test initializing the K8s cluster via Ansible
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Initialize cluster controller
    ${output}=                 Kubectl                    get nodes
    Should Not Match Regexp    ${output}                  ^NAME
    ${rc}=                     Execute Become Playbook    init_cluster.yml
    Should Be Equal            ${rc}                      ${0}
    Wait Until All Nodes Ready
    ${output}=                 Kubectl                    get nodes
    Should Match Regexp        ${output}            \\s+Ready\\s+control-plane

