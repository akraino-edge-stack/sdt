*** Settings ***
Documentation           Check that the crowded notification is recieved
Resource                ../../common.resource
Suite Setup             Log In To Deploy Server
Suite Teardown          Close All Connections

*** Test Cases ***
Check Crowded Notification Is Recieved
    ${output}=             Kubectl              get ep -A | grep edgex 
    Should Match Regexp    ${output}            ^default
    @{lines}=              Split To Lines       ${output}          0
    FOR                    ${line}              IN                 @{lines}
                           ${ip}=               Execute Command    echo "${line}" | awk -F "[ :,]+" '{print $3}'
                           ${output}=           Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59860/api/v2/notification/category/daily-notify'
                           Should Contain       ${output}          "crowding detected"
    END