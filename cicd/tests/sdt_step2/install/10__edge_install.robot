*** Settings ***
Documentation     Test setup of edge nodes
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Setup Edge Nodes
    # TODO: Make sure the edge nodes are not already setup?
    ${rc}=                     Execute Playbook           edge_install.yml
    Should Be Equal            ${rc}                      ${0}
    # Make sure the default runtime of edge node#1 is nvidia.
    ${output}=                 Execute Command            ssh -i ${EDGE_KEY} ${EDGE_USER}@${EDGE_HOST1} docker info
    Should Contain             ${output}                  Default Runtime: nvidia
    # Make sure the default runtime of edge node#2 is nvidia.
    ${output}=                 Execute Command            ssh -i ${EDGE_KEY} ${EDGE_USER}@${EDGE_HOST2} docker info
    Should Contain             ${output}                  Default Runtime: nvidia
    # TODO: Check the installed packages on the edge nodes (script?)
