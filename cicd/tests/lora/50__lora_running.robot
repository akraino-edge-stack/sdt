*** Settings ***
Documentation     Check that the LoRa device is working
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Check device-lora Is Running
    ${output}=             Execute Command        kubectl get pods --namespace=default -o=custom-columns=NAME:.metadata.name,CONTAINERS:.status.containerStatuses[*].name
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN                     @{lines}
                           ${output}=             Split String           ${line}
                           Should Contain         ${output}[1]           edgex-device-lora
    END
Receive Data Via LoRa 
    ${output}=             Execute Command    kubectl get pod -o custom-columns=NODE:.spec.nodeName
    Should Match Regexp    ${output}          ^NODE
    @{lines}=              Split To Lines     ${output}          1
    FOR                    ${line}            IN                 @{lines}
                           ${node}=           Set Variable       ${line}
                           # FIXME: User/password are raw data
                           ${output}=         Execute Command    mosquitto_sub -h localhost -t edgex-events-${node} -u edge -P edgemqtt -W 60
                           Should Contain     ${output}          LoRa-Device
    END
