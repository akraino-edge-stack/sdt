*** Settings ***
Documentation     Test starting EdgeX application via Ansible
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Start EdgeX
    ${rc}=                     Execute Become Playbook    edgex_start.yml
    Should Be Equal            ${rc}                      ${0}
Check All Pods Are Running
    Sleep                      20s
    ${output}=                 Kubectl             get pods
    Should Match Regexp        ${output}           ^NAME
    Should Contain             ${output}           edgex
    Wait Until All Pods Running
Ping All Pods
    ${output}=             Kubectl                get pods -o=custom-columns=NAME:.metadata.name,IP:.status.podIP
    Should Match Regexp    ${output}              ^NAME
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN                     @{lines}
                           ${output}=             Split String           ${line}
                           ${ping}                Execute Command        ssh master ping ${output}[1] -c 5
                           Should Contain         ${ping}                0% packet loss
    END                
Check All Containers Are Running
    ${output}=             Kubectl                get pods --namespace=default -o=custom-columns=NAME:.metadata.name,CONTAINERS:.status.containerStatuses[*].name
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN                     @{lines}
                           ${output}=             Split String           ${line}
                           Should Contain         ${output}[1]           edgex-app-mqtt-export
                           Should Contain         ${output}[1]           edgex-app-rules-engine
                           Should Contain         ${output}[1]           edgex-core-command
                           Should Contain         ${output}[1]           edgex-core-consul
                           Should Contain         ${output}[1]           edgex-core-data
                           Should Contain         ${output}[1]           edgex-core-metadata
                           Should Contain         ${output}[1]           edgex-kong
                           Should Contain         ${output}[1]           edgex-kong-db
                           Should Contain         ${output}[1]           edgex-kuiper
                           Should Contain         ${output}[1]           edgex-redis
                           Should Contain         ${output}[1]           edgex-security-bootstrapper
                           Should Contain         ${output}[1]           edgex-security-proxy-setup
                           Should Contain         ${output}[1]           edgex-security-secretstore-setup
                           Should Contain         ${output}[1]           edgex-support-notifications
                           Should Contain         ${output}[1]           edgex-support-scheduler
                           Should Contain         ${output}[1]           edgex-sys-mgmt-agent
                           Should Contain         ${output}[1]           edgex-vault
                           Should Contain         ${output}[1]           edgex-device-rest
                           Should Contain         ${output}[1]           sync-app
                           Should Contain         ${output}[1]           edgex-device-camera
                           Should Contain         ${output}[1]           image-app
    END
