*** Settings ***
Documentation     Test stopping the K8s cluster via Ansible
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Stop cluster controller
    ${output}=                 Kubectl                    get nodes
    Should Match Regexp        ${output}                  ^NAME
    ${lc}=                     Get Line Count             ${output}
    Should Be Equal            ${lc}                      ${2}                 msg='Only the controller node should be configured'
    ${rc}=                     Execute Become Playbook    reset_cluster.yml
    Should Be Equal            ${rc}                      ${0}
    Sleep                      15s
    ${output}=                 Kubectl                    get nodes
    Should Not Match Regexp    ${output}                  ^NAME

