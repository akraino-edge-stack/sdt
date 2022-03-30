*** Settings ***
Documentation     Test setup of master node
Resource          ../common.resource
Suite Setup       Log In To Deploy Server
Suite Teardown    Close All Connections

*** Test Cases ***
Setup Master Node
    # TODO: Make sure the master node is not already setup?
    ${rc}=                 Execute Become Playbook    master_install.yml
    Should Be Equal        ${rc}                      ${0}
    ${output}=             Execute Command            apt list --installed
    Should Contain         ${output}                  docker.io
    Should Contain         ${output}                  mosquitto
    Should Contain         ${output}                  kubectl
    Should Contain         ${output}                  kubelet
    Should Contain         ${output}                  kubeadm
    File Should Exist      /opt/lfedge/config.yml
    File Should Exist      /opt/lfedge/kube-flannel.yml
    File Should Exist      /opt/lfedge/kube-flannel-private.yml
