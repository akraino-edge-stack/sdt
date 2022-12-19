*** Settings ***
Library          SSHLibrary
Library          String

*** Variables ***
${HOME}             /home/sdt-admin
${DEPLOY_HOST}      sdt-deploy
${DEPLOY_USER}      sdt-admin
${DEPLOY_KEY}       ${HOME}/.ssh/lfedge_deploy
${DEPLOY_PWD}       password
${PLAYBOOK_PATH}    lf-edge/deploy/playbook
${EDGE_HOST1}       jet03
${EDGE_HOST2}       jet04
${EDGE_USER}        edge
${EDGE_KEY}         ${HOME}/.ssh/edge

*** Keywords ***
Log In To Deploy Server
    [Documentation]          Connect to the deploy server over SSH
    Open Connection          ${DEPLOY_HOST}
    Login With Public Key    ${DEPLOY_USER}    ${DEPLOY_KEY}

Execute Playbook
    [Documentation]    Execute an Ansible playbook
    [Arguments]        ${playbook}        ${become}=False    ${path}=${EMPTY}
    ${become_pwd}=     Set Variable If    ${become}
    ...                '-e ansible_become_password=${DEPLOY_PWD}'
    ...                ${EMPTY}
    ${rc}=             Execute Command
    ...                cd ${PLAYBOOK_PATH}${path} ; ansible-playbook -i ./hosts ${become_pwd} ${playbook}
    ...                return_stdout=False
    ...                return_rc=True
    ...                output_during_execution=True
    [Return]           ${rc}

Execute Become Playbook
    [Documentation]    Execute an Ansible playbook with become password
    [Arguments]        ${playbook}        ${path}=${EMPTY}
    ${rc}=             Execute Playbook   ${playbook}    True    ${path}
    [Return]           ${rc}

Execute CICD Playbook
    [Documentation]    Execute an Ansible playbook in the CICD directory
    [Arguments]        ${playbook}        ${become}=False
    ${rc}=             Execute Playbook   ${playbook}    ${become}
    ...                '/../../cicd/playbook'
    [Return]           ${rc}

Execute CICD Become Playbook
    [Documentation]    Execute a CICD Ansible playbook with become password
    [Arguments]        ${playbook}
    ${rc}=             Execute CICD Playbook   ${playbook}    True
    [Return]           ${rc}

Kubectl
    [Documentation]    Execute a kubectl command on master
    [Arguments]        ${options}
    ${output}=         Execute Command    ssh master kubectl ${options}
    [Return]           ${output}

All Nodes Ready
    [Documentation]        Confirm that all K8s nodes are ready
    ${output}=             Kubectl                get nodes
    Should Match Regexp    ${output}              ^NAME
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN            @{lines}
                           Should Match Regexp    ${line}       \\s+Ready\\s+
    END

Wait Until All Nodes Ready
    [Documentation]    Wait until All Nodes Ready succeeds or a 4min timeout
    Wait Until Keyword Succeeds    4 min    15 s    All Nodes Ready

All Pods Running
    [Documentation]        Confirm that all pods in default namespace are running
    ${output}              Kubectl                get pods
    Should Match Regexp    ${output}              ^NAME
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN            @{lines}
                           Should Match Regexp    ${line}       \\s+Running\\s+
    END

Wait Until All Pods Running
    [Documentation]    Wait until All Pods Running succeeds or a 4min timeout
    Wait Until Keyword Succeeds    4 min    15 s    All Pods Running

Execute Commands
    [Documentation]    Execute shell commands
    [Arguments]        ${options}
    ${rc}=             Execute Command
    ...                ${options}
    ...                return_stdout=False
    ...                return_rc=True
    ...                output_during_execution=True
    [Return]           ${rc}