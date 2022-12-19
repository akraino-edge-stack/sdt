*** Settings ***
Documentation           Check that the camera device is working
Resource                ../../common.resource
Suite Setup             Log In To Deploy Server
Suite Teardown          Close All Connections

*** Test Cases ***
Check device-camera Is Running
    ${output}=             Kubectl                get pods --namespace=default -o=custom-columns=NAME:.metadata.name,CONTAINERS:.status.containerStatuses[*].name
    @{lines}=              Split To Lines         ${output}     1
    FOR                    ${line}                IN                     @{lines}
                           ${output}=             Split String           ${line}
                           Should Contain         ${output}[1]           edgex-device-camera
    END
Receive Data(OnvifSnapshot) Via Camera 
    ${output}=             Kubectl            get pod -o custom-columns=NODE:.spec.nodeName
    Should Match Regexp    ${output}          ^NODE
    @{lines}=              Split To Lines     ${output}          1
    FOR                    ${line}            IN                 @{lines}
                           ${node}=           Set Variable       ${line}
                           ${output}=         Execute Command    ssh master mosquitto_sub -h localhost -t edgex-events-${node} -u edge -P edgemqtt -W 60
                           Should Contain     ${output}          "resourceName":"OnvifSnapshot"
    END
Check Readings Of Resource OnvifSnapshot
    ${output}=             Kubectl              get ep -A | grep edgex 
    Should Match Regexp    ${output}            ^default
    @{lines}=              Split To Lines       ${output}           0
    FOR                    ${line}              IN                  @{lines}
                           ${node}=             Execute Command     echo "${line}" | awk -F "[ :,]+" '{print $2}'
                           ${ip}=               Execute Command     echo "${line}" | awk -F "[ :,]+" '{print $3}'
                           IF                   '${node}' == 'edgex-jet03'
                                                ${output}=          Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59880/api/v2/reading/device/name/Camera001/resourceName/OnvifSnapshot'
                                                Should Contain      ${output}          "statusCode":200
                                                Should Not Contain  ${output}          "totalCount":0
                           ELSE
                                                ${output}=          Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59880/api/v2/reading/device/name/Camera002/resourceName/OnvifSnapshot'
                                                Should Contain      ${output}          "statusCode":200
                                                Should Not Contain  ${output}          "totalCount":0
                           END
    END
Receive Shared Data(jpeg) From Other Edge 
    ${output}=             Kubectl            get pod -o custom-columns=NODE:.spec.nodeName
    Should Match Regexp    ${output}          ^NODE
    @{lines}=              Split To Lines     ${output}          1
    FOR                    ${line}            IN                 @{lines}
                           ${node}=           Set Variable       ${line}
                           ${output}=         Execute Command    ssh master mosquitto_sub -h localhost -t edgex-events-${node} -u edge -P edgemqtt -W 60
                           Should Contain     ${output}          "resourceName":"jpeg"
    END
Check The Existence Of Shared Data
    ${output}=             Kubectl              get ep -A | grep edgex 
    Should Match Regexp    ${output}            ^default
    @{lines}=              Split To Lines       ${output}          0
    FOR                    ${line}              IN                 @{lines}
                           ${ip}=               Execute Command    echo "${line}" | awk -F "[ :,]+" '{print $3}'
                           ${output}=           Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59880/api/v2/reading/device/name/sample-image/resourceName/jpeg'
                           Should Contain       ${output}          "statusCode":200
                           Should Not Contain   ${output}          "totalCount":0
    END
Receive Analysis Results(Crowded) From image-app 
    ${output}=             Kubectl            get pod -o custom-columns=NODE:.spec.nodeName
    Should Match Regexp    ${output}          ^NODE
    @{lines}=              Split To Lines     ${output}          1
    FOR                    ${line}            IN                 @{lines}
                           ${node}=           Set Variable       ${line}
                           ${output}=         Execute Command    ssh master mosquitto_sub -h localhost -t edgex-events-${node} -u edge -P edgemqtt -W 60
                           Should Contain     ${output}          "resourceName":"Crowded"
    END
Check The Existence Of Image Analysis Results
    ${output}=             Kubectl              get ep -A | grep edgex 
    Should Match Regexp    ${output}            ^default
    @{lines}=              Split To Lines       ${output}            0
    FOR                    ${line}              IN                   @{lines}
                           ${node}=             Execute Command      echo "${line}" | awk -F "[ :,]+" '{print $2}'
                           ${ip}=               Execute Command      echo "${line}" | awk -F "[ :,]+" '{print $3}'
                           IF                   '${node}' == 'edgex-jet03'
                                                ${output}=           Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59880/api/v2/reading/device/name/Camera001/resourceName/Crowded'
                                                Should Contain       ${output}          "statusCode":200
                                                Should Not Contain   ${output}          "totalCount":0
                           ELSE
                                                ${output}=           Execute Command    ssh master curl --noproxy ${ip} 'http://${ip}:59880/api/v2/reading/device/name/Camera002/resourceName/Crowded'
                                                Should Contain       ${output}          "statusCode":200
                                                Should Not Contain   ${output}          "totalCount":0
                           END
    END
