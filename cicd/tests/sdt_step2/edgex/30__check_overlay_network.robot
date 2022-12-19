*** Settings ***
Documentation     Test Ping Via Overlay Network
Resource          ../../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Check Ping Via Overlay Network
    ${output}=             Kubectl            get pod -o=custom-columns=NAME:.metadata.name,IP:.status.podIP
    Should Match Regexp    ${output}          ^NAME
    @{lines}=              Split To Lines     ${output}
    ${line}                Set Variable       ${lines}[1]
    ${output}=             Split String       ${line}
    ${pod_1}=              Set Variable       ${output}[0]
    ${line}                Set Variable       ${lines}[2]
    ${output}=             Split String       ${line}
    ${pod_2_url}=          Set Variable       ${output}[1]
    # NOTE: ping will not work with security fixes
    ${output}=             Execute Command    ssh master kubectl exec -it ${pod_1} -c sync-app -- wget http://${pod_2_url}:8500 -O output    return_stdout=False    return_stderr=True
    Should Contain         ${output}          'output' saved
