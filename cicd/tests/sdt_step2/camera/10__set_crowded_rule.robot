*** Settings ***
Documentation           Set crowded notification rule
Resource                ../../common.resource
Suite Setup             Log In To Deploy Server
Suite Teardown          Close All Connections

*** Test Cases ***
Set Crowded Rule
    ${rc}=              Execute Commands    chmod +x ${PLAYBOOK_PATH}/../../edgex/rules/set-crowded-notification-rule.sh
    Should Be Equal     ${rc}               ${0}
    ${rc}=              Execute Commands    scp ${PLAYBOOK_PATH}/../../edgex/rules/set-crowded-notification-rule.sh master:/tmp
    Should Be Equal     ${rc}               ${0}
    ${rc}=              Execute Commands    ssh master /tmp/set-crowded-notification-rule.sh
    Should Be Equal     ${rc}               ${0}
    ${rc}=              Execute Commands    ssh master rm /tmp/set-crowded-notification-rule.sh
    Should Be Equal     ${rc}               ${0}