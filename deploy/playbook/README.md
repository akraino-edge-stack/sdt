# Deploy Playbooks

Ansible must be installed on the deploy node where these scripts will be used.

```
sudo add-apt-repository ppa:ansible/ansible
sudo apt-get install ansible
```

NOTE: The version of Ansible available from the Ubuntu repository for 20.04
does not support the kubernetes.core collection used by some of these scripts.
This is why the Ansible respository is added in the command above.

Ansible scripts (called playbooks) declare the desired end state
of the process, and only perform steps that are required to change the
current state. This means that, with few exceptions, it is safe to rerun
playbooks as they will not change anything that is not required.

In addition, the `community.docker` and `kubernetes.core` collections for
Ansible need to be installed on the deploy node. This is automatically
done by the `setup_cicd.yml` playbook in `cicd/playbook`, or can be done
using the following commands (executed by the user who will run the deploy
playbooks on the deploy node) after Ansible is installed.

```
ansible-galaxy collection install community.docker
ansible-galaxy collection install kubernetes.core
```

## Prepare nodes for deployment

The admin user must have sudo privileges on the edge nodes. Set the
sudo password in the `deploy/playbook/group_vars/edge_nodes/secret` file.

Add the node names to the `hosts` file in this directory, under the
`edge_nodes` group. The example below configures one edge node named "edge".
Take care to include the colon after the host name.

```
    edge_nodes:
      hosts:
        edge:
          ip: 192.168.2.14
```

Add IP addresses to the hosts file for each edge node as shown above.
These will be added to the master node's `/etc/hosts` file by the master
install script (if they are not already present). Also, set the master
IP address appropriately in the file `deploy/playbook/group_vars/all`.

```
master_ip: 192.168.2.10
```

It may be useful to run the `master_install.yml` playbook at this point to
update `/etc/hosts` before performing the next step.

Create an admin user key to allow passwordless ssh for all your edge nodes.
The example below assumes a user "edge" exists on all the nodes. The commands
create and copy keys for the edge user to a node named "edge".

```
ssh-keygen -t ed25519 -f ~/.ssh/edge
ssh-copy-id -i ~/.ssh/edge.pub edge@edge
```

Repeat the `ssh-copy-id` line for all nodes. Test by using
`ssh -i ~/.ssh/edge edge@edge` to login.

In the `vars` section in the `hosts` file for `edge_nodes` set the
`ansible_user` variable to the correct admin user name to be used to log in
to hosts, and set the `ansible_ssh_private_key_file` to the location of the
file specified in the `ssh-keygen` command above.

The `master_install.yml` playbook assumes that the deploy scripts will be
run on the master node by an admin user with sudo priviledges. Use the
`--ask-become-pass` option when running it to supply the sudo password. If
the master and deploy nodes are separate, replace localhost in the hosts
file in the `master` group, and remove the `connection: local` line in
`master_install.yml`. You may also add handling for login credentials (as
in `edge_install.yml` for example).

## Install Required Packages

Package installation uses apt, and assumes the standard repositories are
reachable from all nodes.

### Master

Run the `master_install.yml` playbook to install required packages on the
master node and perform other initialization.

```
ansible-playbook -i ./hosts master_install.yml --ask-become-pass
```

NOTE: An entry will be added to `/etc/hosts` to map the host name "master"
to the local host (if that is not already its host name).

### Edge

The `deploy/playbook/group_vars/all` file defines a variable
`master_ip` with the IP address of the master node. Make sure this
matches your network configuration. The install playbooks will
use this variable to add an entry to each node's `/etc/hosts` file.

Run the `edge_install.yml` playbook to install required packages on all
edge nodes.

```
ansible-playbook -i ./hosts edge_install.yml
```

## Docker Registry Management

Use the playbooks `start_registry.yml`, `stop_registry.yml`, and
`remove_registry.yml` to manage the local Docker registry on the master
node. The registry runs as a docker container named "registry" and is
accesible through port 5000. The local registry will be configured to
restart when docker is restarted, but not if it is stopped using
`stop_registry.yml`, in which case it will need to be restarted with the
`start_registry.yml` playbook.

The start registry playbook needs to be run once to start the registry on
the master node. Once started, the registry will remain running as a service
even if the node reboots.

```
ansible-playbook -i ./hosts start_registry.yml --ask-become-pass
```

NOTE: The start registry playbook also creates certificates to secure
communication between the edge nodes and the registry, which need to be 
copied to the edge nodes. For this reason, the start registry script
needs to be run at least once before the `edge_install.yml` script,
which will copy those certificates to the edge nodes.

If the certificates need to be copied to an edge node manually, copy
the files `/etc/docker/certs.d/master:5000/registry.crt` and
`/usr/local/share/ca-certificates/master.crt` from the master node to
the same locations on the edge node, and run the following command on
the edge node.

```
sudo update-ca-certificates
```

The local Docker registry can be populated with the required images
using the `pull_upstream_images.yml` playbook. This script should be run
after the local registry has been started. The list of EdgeX images 
is determined from the configuration files in the `edgex` subdirectories.
The required images for Kubernetes and Flannel are automatically determined.

```
ansible-playbook -i ./hosts pull_upstream_images.yml
```

This script will pull the images from Docker Hub, and the Kubernetes
and Flannel default registries, and push them to the local registry service.

The `clean_local_images.yml` playbook will remove the upstream images from
the local host. It will not remove the images from the local registry.
To clean the local registry, use the `remove_registry.yml` playbook, which
will stop the registry service and remove files from the master host.

It is possible to confirm the names of images currently stored in the local
registry with the following command:

```
curl http://master:5000/v2/_catalog
```

## Cluster Configuration

The cluster configuration files will be installed in `~/.lfedge` by the
master install script.

The cluster can be initialized using the `init_cluster.yml` playbook:

```
ansible-playbook -i ./hosts init_cluster.yml --ask-become-pass
```

NOTE: This will reset the cluster if it is already running. It relies on
the master install and pull upstream images playbooks having executed
successfully.

The `init_cluster.yml` playbook does not configure or start any edge nodes
at this time.

The cluster can be reset (stopped) with the `reset_cluster.yml` playbook.

To add edge nodes to the cluster, execute the `join_cluster.yml` playbook.

```
ansible-playbook -i ./hosts join_cluster.yml
```

You can confirm the cluster is up and running using the `kubectl` command on
the master node. Note that it may take some time for the edge nodes to become
ready after the join script has been run. In the example below "ubuntu20" is
the master node.

```
colin@ubuntu20:~/lf-edge/deploy/playbook$ kubectl get nodes
NAME       STATUS     ROLES                  AGE   VERSION
edge       NotReady   <none>                 28s   v1.22.4
ubuntu20   Ready      control-plane,master   14m   v1.22.4
colin@ubuntu20:~/lf-edge/deploy/playbook$ kubectl get nodes
NAME       STATUS   ROLES                  AGE     VERSION
edge       Ready    <none>                 4m19s   v1.22.4
ubuntu20   Ready    control-plane,master   18m     v1.22.4
colin@ubuntu20:~/lf-edge/deploy/playbook$
```

The playbook `delete_from_cluster.yml` can be used to remove all edge nodes
from the cluster (i.e. reverse the effect of `join_cluster.yml`). The master
node will remain active.

```
ansible-playbook -i ./hosts delete_from_cluster.yml
```

To completely shut down the cluster run `delete_from_cluster.yml` followed
by `reset_cluster.yml`.

### EdgeX Services

To start the EdgeX services on the Edge nodes, run the `edgex_start.yml`
playbook.

```
ansible-playbook -i ./hosts edgex_start.yml
```

The script creates configuration files under `~/.lfedge/<node>` for the
Edge nodes and adds them to the Kubernetes cluster configuration. The
cluster must be initialized (e.g. using `init_cluster.yml` and
`join_cluster.yml`) and the Docker registry must be ready (e.g. started using
`start_registry.yml` and `pull_upstream_images.yml`).

To stop the EdgeX services and delete the configuration from the cluster,
run the `edgex_stop.yml` playbook.

```
ansible-playbook -i ./hosts edgex_stop.yml
```

#### Enabling EdgeX Device Services

Variables in the `edgex_start.yml` playbook control which device
services are installed in the cluster, the variables are boolean values
(true or false). If set to true, the corresponding device service is
installed. Change the set of installed services on the cluster by first
stopping EdgeX, if necessary, with the `edgex_stop.yml` script, and
then modifying the `edgex_start.yml` script before running it again to
start EdgeX.

The device service variables are as follows:

* `device_virtual`: Enables the device-virtual service provided by EdgeX,
for testing purposes.
* `device_lora`: Enables the temperature sensor reading via a LoRa transport
service (see `edgex/device-lora` for details).

At this time the device-rest service is always installed to
provide the ability to exchange data between nodes.

## Summary of Setup Sequence

All steps are performed on the deploy node unless otherwise noted. Playbooks
that require the `--ask-become-pass` option will prompt for the sudo password
of the master node admin user (the user running the playbook).

1. Install Ansible: `sudo add-apt-repository ppa:ansible/ansible && sudo apt-get install ansible`
1. Add the `community.docker` and `kubernetes.core` Ansible collections: `ansible-galaxy collection install community.docker` and `ansible-galaxy collection install kubernetes.core` (or run `ansible-playbook -i ./hosts setup_cicd.yml --ask-become-pass` in `cicd/playbook`, which will also install Jenkins and Robot Framework).
1. Set the edge node admin user's sudo password in `deploy/playbook/group_vars/edge_nodes/secret`.
1. Add the edge nodes and their IP addresses to the `deploy/playbook/hosts` file.
1. Set the master node's IP address in the `deploy/playbook/group_vars/all` file.
1. Install and configure software for the master node: `ansible-playbook -i ./hosts master_install.yml --ask-become-pass`
1. Start the local Docker registry: `ansible-playbook -i ./hosts start_registry.yml --ask-become-pass`
1. Populate the local Docker registry with required images: `ansible-playbook -i ./hosts pull_upstream_images.yml`
1. Configure passwordless ssh for the edge node admin user to all edge nodes: Run `ssh-keygen -t ed25519 -f ~/.ssh/edge`, then run `ssh-copy-id -i ~/.ssh/edge.pub <edge-node-admin-user>@<edge-node-hostname>` for each edge node.
1. Set the edge node admin user name in the `ansible_user` variable in the `vars` section of the `edge_nodes` group in `deploy/playbook/hosts` (optionally change the `ansible_ssh_private_key_file` variable if you did not use `~/.ssh/edge`).
1. Install and configure software for the edge nodes: `ansible-playbook -i ./hosts edge_install.yml`
1. Initialize the master node as the Kubernetes cluster controller: `ansible-playbook -i ./hosts init_cluster.yml --ask-become-pass`
1. Add all edge nodes to the cluster: `ansible-playbook -i ./hosts join_cluster.yml`
1. Start the EdgeX services: `ansible-playbook -i ./hosts edgex_start.yml`

## Script Reference

| Task | Command |
|----|----|
| Setup master node | `ansible-playbook -i ./hosts master_install.yml --ask-become-pass` |
| Setup edge nodes | `ansible-playbook -i ./hosts edge_install.yml` |
| Start Docker registry | `ansible-playbook -i ./hosts start_registry.yml --ask-become-pass` |
| Populate Docker registry with required images | `ansible-playbook -i ./hosts pull_upstream_images.yml` |
| Remove cached Docker images from master | `ansible-playbook -i ./hosts clean_local_images.yml` |
| Stop Docker registry | `ansible-playbook -i ./hosts stop_registry.yml` |
| Remove Docker registry | `ansible-playbook -i ./hosts remove_registry.yml` |
| Create K8s cluster with master as controller | `ansible-playbook -i ./hosts init_cluster.yml --ask-become-pass` |
| Add edge nodes to K8s cluster | `ansible-playbook -i ./hosts join_cluster.yml` |
| Remove edge nodes to K8s cluster | `ansible-playbook -i ./hosts delete_from_cluster.yml` |
| Shut down K8s cluster controller | `ansible-playbook -i ./hosts reset_cluster.yml --ask-become-pass` |
| Start EdgeX services on edge nodes | `ansible-playbook -i ./hosts edgex_start.yml` |
| Stop EdgeX services and delete configuration | `ansible-playbook -i ./hosts edgex_stop.yml` |

