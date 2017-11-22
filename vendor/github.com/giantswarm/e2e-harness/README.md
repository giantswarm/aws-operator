# e2e-harness

[![CircleCI](https://circleci.com/gh/giantswarm/e2e-harness.svg?style=shield)](https://circleci.com/gh/giantswarm/e2e-harness)

Harness for custom kubernetes e2e testing.

## Getting Project

Clone the git repository: https://github.com/giantswarm/e2e-harness.git

## Running e2e-harness

You can download a prebuilt binary from [here](https://github.com/giantswarm/e2e-harness/releases/) or,
with a golang environment set up, build from source from the root of the project:
```
go build .
```

## How does e2e-harness work

The goal of the project is making it easy to design and run e2e tests for kubernetes
components. We have great tools for local development, like minikube, but the tests
and results obtained using them are difficult to replicate on a CI environment.

e2e-harness aims to abstract all the differences between local and CI environments,
so that you can write the tests once and run them everywhere, making sure that if
things work locally they will work too elsewhere.

In order to achive that, e2e-harness has two operation modes: local and remote.
The setup and teardown actions differ on each mode, but the test themselves (and
the actions required to execute them) are the same.

Regarding the test execution, all the actions are run on a container, so that the
execution environment is always the same. We have put in place these binaries inside
the container:

* kubectl: k8s CLI client, allows us to run common setup actions or out of cluster
tests (see below).
* helm: at giantswarm most of the systems under tests are helm charts. We have
installed the registry plugin too.
* shipyard: it allows us to create and delete a remote minikube instance.

The container image is published in quay registry `quay.io/giantswarm/e2e-harness:latest`
and its Dockerfile can be found [here](https://github.com/giantswarm/e2e-harness/blob/master/Dockerfile).

## Requirements

The main requirement is having a recent docker version running on the host. Additionally,
for each operation mode:

* Local: [minikube](https://github.com/kubernetes/minikube) should be started with
RBAC enabled before running e2e-harness:

```
$ minikube start --extra-config=apiserver.Authorization.Mode=RBAC
```

* Remote: as stated above e2e-harness uses [shipyard](https://giathub.com/giantswarm/shipyard)
for setting up the remote cluster. shipyard currently only supports AWS as the
backend engine, so the common environment variables for granting access to AWS
are required too (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`).

## e2e-harness lifecycle

An e2e test execution involves three stages:

* Setup: this is performed with the `setup` command.

```
$ e2e-harness setup --remote=[false|true]
```

It takes care of:
  - Initializing the project, creating an interchange directory that will
  be mounted on the test container for keeping state between command
  executions.
  - Prepare the connection to the test cluster: for remote executions this
  will involve the creation of the cluster too. In local executions, the
  connection settings to be able running minikube are made available in
  the test container interchange directory.
  - Run common set up steps: these will put the test cluster in a common
  initial state, basically installing tiller (helm's server side part) including
  the required resources to make it work with RBAC enabled.
  - Run specific setup steps: this are defined in the project file
  `e2e/project.yaml` under the `setup` key. They are common steps (see their
  description above) and can do things like installing the chart under test,
  setting up required external resources, etc.

* Run tests: invoking the `test` command:

```
$ e2e-harness test
```

First, the binary and docker image for the project are built. Then the tests
are compiled, not executed directly. Finally, that binary is executed in the
test container with an environment defined by the environment variables set
in the `project.yaml` file.

* Teardown: the teardown phase is executing using the `teardown` command.

```
$ e2e-harness teardown
```

It consists of running common tear down steps: these differ depending on the mode
of operation, for remote ephemeral clusters, they are just deleted, for local
clusters tiller and all the required RBAC setup is removed.

## Project initialization

From the project root execute:

```
$ e2e-harness init
```

This will create a `e2e` directory, with the required files to start writing tests,
see below. This is how the e2e directory looks like this:

```
├── client.go
├── example_test.go
└── project.yaml
```

`project.yaml` defines how the end to end tests should be setup using environment variables, they
are defined as follows:

```
test:
  env:
  - ENV_VAR_1=VALUE_1
  - ENV_VAR_2=${VALUE_2}
```
In this case, `${VALUE_2}` would be expanded and its value would be available for the tests.

* `client.go`: library with a k8s client.
* `example_test.go`: contains example tests.

## Writing tests

e2e tests are executed from the test container and are regular go test, for writing them
just keep in mind these considerations:

* The tests will be executed from the test container, this is it's [Dockerfile](https://github.com/giantswarm/e2e-harness/blob/master/Dockerfile).
* All the go files should be guarded by a `k8srequired` build tag, being their first line:
```
// +build k8srequired
```
* The kube config file path for connecting to the test cluster can be obtained from
the `DefaultKubeConfig` constant in the `giantswarm/e2e-harness/pkg/harness` package.

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/e2e-harness/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING.md](/giantswarm/e2e-harness/blob/master/CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

E2e-Harness is under the Apache 2.0 license. See the [LICENSE](/giantswarm/e2e-harness/blob/master/LICENSE) file for details.
