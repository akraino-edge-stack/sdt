*** Settings ***
Documentation     Test setup of edge nodes
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Setup Edge Nodes
    # TODO: Make sure the edge nodes are not already setup?
    ${rc}=                     Execute Playbook           edge_install.yml
    Should Be Equal            ${rc}                      ${0}
    # TODO: Check the installed packages on the edge nodes (script?)
