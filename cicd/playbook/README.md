# Setup CI/CD Server

The following command should be run by the user with sudo permissions which
will be running Ansible. As currently configured it will run against the local
host. When prompted for the "become" password, supply the sudo password for
the user's account.
```
ansible-playbook -i ./hosts setup_cicd.yml --ask-become-pass
```

The playbook will perform the following steps:

* Install Basic Build Tools
* Install Jenkins (and the Java runtime it requires)
* Add build server to file /etc/hosts 

## Jenkins Configuration

After Jenkins is first installed, follow the
[instructions here](https://www.jenkins.io/doc/book/installing/linux/#setup-wizard)
to complete the initial setup.

If you are running behind a proxy server do not forget to configure the
proxy settings in the "Manage Jenkins", "Manage Plugins" screen under the
"Advanced" tab. These settings (including the "No Proxy" list) also affect
some plugins, such as the Git plugin.

Jenkins requires some additional configuration to support running the build
and test scripts. See the README file in `cicd/jenkins` for details.

The `import_jobs.yml` playbook will import job configurations into Jenkins
to perform builds and tests. This only needs to be done once unless the
job configurations are changed upstream.

```
ansible-playbook -i ./hosts import_jobs.yml --ask-become-pass
```

The `export_jobs.yml` file can be used to export ***all*** job configurations
from Jenkins into the `cicd/jenkins/jobs` folder, to support development of
new test cases. You will not need to run this script unless you are writing
or modifying builds or test cases and Jenkins jobs to run them. Note that
this only copies the job configuration, not any credentials or existing build
records.

```
ansible-playbook -i ./hosts export_jobs.yml
```

NOTE: Jenkins may not display reports and logs from the Robot Framework
properly (showing that javascript is disabled) without extra configuration.
One potential workaround is to use the
[Resource Root URL](https://www.jenkins.io/doc/book/security/user-content/#resource-root-url)
setting, in the "Manage Jenkins", "Serve resource files from another domain"
section. If you are using the system in a small or closed network, you
can assign a second host name to the CI/CD server running Jenkins and
use that name in the Resource Root URL, while using the primary host name
in the "Jenkins URL" setting. This works around the default "Content Security
Policy" which disables javascript when serving "resources", such as build
results.

# Application Build Environment

In order to build the custom application and device services, a build
environment needs to be set up using `setup_build.yml`. This can be done on
the CI/CD or deploy servers, or on a separate machine (but must be on an
`x86_64` architecture). In this release, the build environment is on 
a seprate machine. The build server must have access to the local
Docker registry (see `deploy/playbook/README.md`). The Docker registry
name should be configured in `deploy/playbook/group_vars/all`. Run the
setup script on the deploy server:

```
ansible-playbook -i ./hosts setup_build.yml --ask-become-pass
```

The playbook will perform the following steps:

* Make sure there is an entry for the master node and deploy node in /etc/hosts
* Install required software packages including Docker and Go and Robot Framework
* Make sure the user can run Docker commands
* Configure Docker, including adding the certificates to secure access to the private registry
* Create the `edgexfoundry` directory

NOTE: The Robot Framework plugin should also be added to Jenkins via the
Jenkins management interface.

## ARM Build Server

The `build_images.yml` and `push_images.yml` scripts discussed below build
all images on the x86 build server, using emulation and Docker to build the
ARM images. This is the recommended option, but it is also possible to build
the ARM images using a second server configured using the `setup_arm_build.yml`
script.

Using one of the edge nodes for the ARM build server should work, as long
as the node has sufficient resources. The ARM build server must also have
access to the local Docker registry. If the ARM build server is configured
as an edge node using `deploy/playbook/edge_install.yml`, the registry host
name should already be in the `/etc/hosts` directory.

Specify the server and account to be used to build ARM applications with the
`cicd/playbook/hosts` file, under the `arm-build` host:

```
    arm-build:
      ansible_host: erc01
      ansible_user: edge
      ansible_ssh_private_key_file: ~/.ssh/edge
      ansible_become_password: password
```

`ansible_host` is the host name of the ARM build server.

The SSH private key file must be set up for accessing the build account in
the same way as the admin account for edge nodes described in
`deploy/playbook/README.md`.

After the configuration above, you can install the ARM build server by running 
the following command on the deploy server.
```
ansible-playbook -i ./hosts setup_arm_build.yml
```

## Build And Push Images

The custom applications can be built using the `build_images.yml`,
`build_amd64.yml` and `build_arm64.yml` scripts. Note that all of these can
be run on the deploy server. The `build_arm64.yml` script will log
in to the ARM build server via SSH, using the settings in
`cicd/playbook/hosts`. The `build_images.yml` script is recommended, as it
does not require a dedicated ARM build server and builds all images. An ARM
server does not need to be configured if only `build_images.yml` and
`push_images.yml` will be used.

Build all images:

```
ansible-playbook -i ./hosts build_images.yml
```

Build x86 (amd64) images only:

```
ansible-playbook -i ./hosts build_amd64.yml
```

NOTE: The build process can take several minutes or more, especially the
first time, as the base images and packages are downloaded.

Build ARM images only, using a separate ARM build server:

```
ansible-playbook -i ./hosts build_arm64.yml
```

Push the completed images to the local Docker registry on the deploy node:

```
ansible-playbook -i ./hosts push_images.yml
```

If ARM devices will not be used, you can push only the x86 images:

```
ansible-playbook -i ./hosts push_amd64_images.yml
```

NOTE: Both x86 and ARM images must be built if the `push_images.yml` script
will be used. Only x86 images are necessary if `push_amd64_images.yml` will
be used.

NOTE: Running `push_amd64_images.yml` will make the local Docker registry
supply only x86 images for the custom applications. Running `push_images.yml`
will provide both x86 and ARM images based on the architecture of the node
pulling the image. Each script will override the results of any previously
run script.

If the source for an application is updated, run the above scripts
(build and push) again to build it and push the images to the Docker
registry.

## Summary Of Build Process

1. Install the build environment: `ansible-playbook -i ./hosts setup_build.yml --ask-become-pass`
1. Build the applications: `ansible-playbook -i ./hosts build_images.yml`
1. Push applications to local Docker registry: `ansible-playbook -i ./hosts push_images.yml`

## Build Script Reference

| Task | Command |
|----|----|
| Install build tools | `ansible-playbook -i ./hosts setup_build.yml --ask-become-pass` |
| Install build tools on ARM server | `ansible-playbook -i ./hosts setup_arm_build.yml` |
| Build applications | `ansible-playbook -i ./hosts build_images.yml` |
| Build x86 applications only | `ansible-playbook -i ./hosts build_amd64.yml` |
| Build ARM applications only | `ansible-playbook -i ./hosts build_arm64.yml` |
| Push applications to local Docker registry | `ansible-playbook -i ./hosts push_images.yml` |
| Push x86/amd64 images only | `ansible-playbook -i ./hosts push_amd64_images.yml` |
