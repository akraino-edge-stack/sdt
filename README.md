# Smart Data Transaction for CPS Blueprint Project

The system consists of the following types of node:

* CI/CD: Runs Jenkins server to control the build node.
* Build: Pull source and scripts, build components and run robot tests.
* Deploy: Runs scripts (mainly Ansible playbooks) to install components on
  master and edge nodes.
* Master: Runs the Kubernetes controller for orchestrating the edge nodes,
  a local docker registry providing container images for the edge nodes,
  and mosquitto (MQTT broker) for collecting data from edge nodes and
  sending commands to them.
* Edge: Collects sensor data, performs edge processing, and forwards data
  to the MQTT broker on the master node.
* Sensor: Not a full-fledged node, but a device containing sensor hardware and
  a communications device (e.g. LoRa) so that sensing data can be collected
  remotely by the edge nodes.
* Camera: IP camera device (e.g. HW-500E6A) so that image data can be collected 
  remotely by the edge nodes.

Initially, the CI/CD, deploy, and master roles are all handled by a single
host, with the host name "master".
In Release7, the CI/CD, build, deploy, and master roles are assigned to 
different nodes. And also some improvements have been made. All nodes can be 
installed from the deploy node.

Scripts and instructions for setting up the CI/CD and build nodes are in the directory
`cicd/playbook`. Scripts used to manage the master and edge nodes and services
(either through CI/CD or manually) are in the directory `deploy/playbook`.
Test scripts are in `cicd/tests`,
and Jenkins build and test job configurations are under `cicd/jenkins`.
See README files in those directories for more details.

Information and scripts for setting up sensor nodes is located in the `sensor`
directory. See the README there for more details.

## Summary of Setup

1. Prepare CI/CD, Build, Deploy, Master, and two or more Edge nodes
1. Install Ansible on the Deploy node
1. Run the `setup_deploy.yml` (in `deploy/playbook`) playbook on deploy node to install deploy software
1. Edit the hosts and variable files in `deploy/playbook` to match your configuration
1. Run the `master_install.yml` playbook to install required software on the Master node
1. Run `start_registry.yml` to initialize the local Docker registry
1. Run `pull_upstream_images.yml` to populate the local Docker registry with required container images
1. Edit the hosts file in `cicd/playbook` to match your configuration
1. Run the `setup_cicd.yml` (in `cicd/playbook`) playbook on deploy node to install CI/CD software
1. Run the `setup_build.yml` (in `cicd/playbook`) playbook on deploy node to install build software
1. Run `build_images.yml` and `push_images.yml` (in `cicd/playbook`) on deploy node to push the custom microservice images to the registry
1. Setup the admin user account for all Edge nodes
1. Run the `edge_install.yml` playbook on deploy node to install required software on the Edge nodes
1. Run `init_cluster.yml` on deploy node to initialize the Kubernetes controller node (Master node)
1. Run `join_cluster.yml` on deploy node to initialize the Kubernetes worker nodes (Edge nodes)
1. Run `edgex_start.yml` on deploy node to start the EdgeX services on the Edge nodes

See the README.md files in `cicd/playbook` and `deploy/playbook` for more
details.
