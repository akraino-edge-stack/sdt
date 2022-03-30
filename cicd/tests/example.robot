*** Settings ***
Documentation    Example test over SSH
Library          OperatingSystem
Library          SSHLibrary
Suite Setup      Open Connection And Log In
Suite Teardown   Close All Connections

*** Variables ***
${HOME}          /home/colin
${DEPLOY_HOST}   localhost
${DEPLOY_USER}   colin
${DEPLOY_KEY}    ${HOME}/.ssh/lfedge_deploy

*** Test Cases ***
Simple Echo Test
    ${output}=         Execute Command    echo Hello, World!
    Should Be Equal    ${output}          Hello, World!

*** Keywords ***
Open Connection And Log In
    [Documentation]          Connect to the deploy server over SSH
    Open Connection          ${DEPLOY_HOST}
    Login With Public Key    ${DEPLOY_USER}    ${DEPLOY_KEY}
