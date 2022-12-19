*** Settings ***
Documentation     Check data is received from edge nodes
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Receive MQTT messages from edge nodes
    ${output}=             Kubectl            get pod -o custom-columns=NODE:.spec.nodeName
    Should Match Regexp    ${output}          ^NODE
    @{lines}=              Split To Lines     ${output}          1
    FOR                    ${line}            IN                 @{lines}
                           ${node}=           Set Variable       ${line}
                           # FIXME: User/password are raw data
                           ${output}=         Execute Command    ssh master mosquitto_sub -h localhost -t edgex-events-${node} -u edge -P edgemqtt -W 60
                           Should Contain     ${output}          profileName
    END
