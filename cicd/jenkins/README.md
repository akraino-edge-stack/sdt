# Jenkins Build and Test Jobs

This directory contains jobs to perform build and test tasks in Jenkins.

The playbooks `import_jobs.yml` and `export_jobs.yml` in the `cicd/playbook`
directory can be used to copy these jobs into Jenkins or export all the
job configurations from Jenkins into this directory.

In order to use these jobs there must be appropriate credentials configured
in Jenkins and a SSH public key installed on the build and deploy server, 
so that Jenkins can log into the build server to run the robot test scripts.

## Required Plugins

The Git, SSH Credentials, and Robot Framework plugins for Jenkins are required.
(Note that the Robot Framework itself is also required. See the README file
in `cicd/tests`.)
They can be installed through the "Manage Jenkins", "Manage Plugins" screen.

## Required Credentials

Configure the following credentials in Jenkins via the "Manage Jenkins",
"Manage Credentials" screen in the "global" domain. The IDs given below must
be used or the jobs will not be able to find the correct credentials.

| ID | Credential Type | Content |
|----|----|----|
| lfedge-deploy-key | SSH Username with private key | The user name and a private key to log in via SSH to the deploy server |
| lfedge-gitlab-login | Username with password | The user name and password used to access the git repository (see below) |
| lfedge-deploy-password | Secret text | The sudo password for the user specified in lfedge-deploy-key |

NOTE: The test scripts default to logging in to localhost as the deploy server
(i.e. the CI/CD server and the deploy server are the same). This can be
changed by editing the `DEPLOY_HOST` variable in the
`cicd/tests/common.resource` file.

## SSH Key

Create a key pair for Jenkins to use to log in to the build and deploy server, 
for example with the following command:

```
ssh-keygen -t rsa -f lfedge_deploy
```

The example above will create two files: `lfedge_deploy` and
`lfedge_deploy.pub`. The content of `lfedge_deploy` should be set as the
private key data in `lfedge-deploy-key` in Jenkins (paste the entire content
of the file, including the `-----BEGIN OPENSSH PRIVATE KEY-----` and
`-----END OPENSSH PRIVATE KEY-----` lines). Add the content of the
`lfedge_deploy.pub` file (a single line beginning with `ssh-rsa` and ending
with `username@hostname`) to the user's `~/.ssh/authorized_keys` on the
build and deploy servers.

## Git Repository

The git repository URL specified in the jobs will need to be changed to match
a reachable git repository with this source distribution. Jenkins will
clone the source when running jobs to access the scripts in the `cicd/tests`
directory.

Click the job name on the Jenkins dashboard and select "configure", then
set the "Repository URL" in the "Source Code Management" section to the
appropriate value and click "Save". This will be overwritten if the job
configuration is imported again. Alternatively, before importing the jobs,
edit the `config.xml` file in each directory under `cicd/jenkins/jobs` and
replace the URL in the `userRemoteConfigs` section:

```
    <userRemoteConfigs>
      <hudson.plugins.git.UserRemoteConfig>
        <url>http://gitlab.falcon.qnet.fujitsu.com/iown/lf-edge.git</url>
        <credentialsId>lfedge-gitlab-login</credentialsId>
      </hudson.plugins.git.UserRemoteConfig>
    </userRemoteConfigs>
```

