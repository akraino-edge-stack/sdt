# Test Scripts

The test scripts in this directory use the
[Robot Framework](https://robotframework.org/)
to execute commands and check results. The Robot Framework SSH library
extension is also required. These components can be installed using pip:

```
sudo pip install robotframework robotframework-sshlibrary
```

The `cicd/playbook/setup_build.yml` playbook does this automatically.

Configure the variables in `common.resource` to match your environment if
you plan to run the test scripts directly. If you are running the scripts
from Jenkins jobs (see `cicd/jenkins`), only the `PLAYBOOK_PATH` and
`DEPLOY_HOST` variables need to be configured, as the other data is overridden
by values supplied from Jenkins.

You can also supply appropriate values from the command line when running
robot, using the `-v` option:

```
robot -v DEPLOY_USER:user -v DEPLOY_KEY:path/to/key -v DEPLOY_PWD:password cluster/10__init_cluster.robot
```

NOTE: Be careful not to commit or otherwise make public the
`common.resource` file if you have configured the `DEPLOY_PWD` variable.
Supplying a password directly on the command line as above can also be a
security risk.

## Test Grouping

Tests are grouped in folders and ordered using "NN`__`" prefixes recognized
by Robot so they can be executed as a suite in sequence to test functionality.
For example, the `cluster` directory tests initializing the K8s cluster,
adding edge nodes to an initialized cluster, removing edge nodes, and
stopping the cluster, in that order.

The test groups are as follows:

* `install`: Install required software and configuration on master and edge nodes
* `docker`: Set up the local Docker registry, update it, clean up and remove it
* `cluster`: Start and stop the K8s cluster, and add and remove edge nodes
* `edgex`: Start and stop the EdgeX microservices, and verify the edge nodes are sending data to Mosquitto when the microservices are running

The test cases make assumptions about the state of the systems under test.
For example, the `cluster` test cases assume the required software is installed
and the Docker registry is running and up to date. The `edgex` test cases
assumes the K8s cluster has master and edge nodes ready.
