[![CircleCI](https://circleci.com/gh/giantswarm/aws-operator.svg?&style=shield&circle-token=8f0fe6ad08c090afa36c35ba5d926ac6ffe797e8)](https://circleci.com/gh/giantswarm/aws-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/aws-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/aws-operator)


# aws-operator

The aws-operator handles Kubernetes clusters running on a Kubernetes cluster
inside of AWS.


## Prerequisites


## Getting Project

Download the latest release:
https://github.com/giantswarm/aws-operator/releases/latest

Clone the git repository: https://github.com/giantswarm/aws-operator.git

Download the latest docker image from here:
https://quay.io/repository/giantswarm/aws-operator


### How to build

This project provides a Makefile, so you can build it by typing:

```
make
```

If you prefer, you may also build it using the standard `go build` command,
like:

```
go build github.com/giantswarm/aws-operator
```


## Running aws-operator

See [this guide][examples-local].

[examples-local]: https://github.com/giantswarm/aws-operator/blob/master/examples/local


## Architecture


### S3 buckets

S3 buckets are used for storing cloudconfigs, which are fetched by the EC2
instances and executed. Then cloud-init, by using those cloudconfigs, takes
care of running the needed systemd services and containers. The result of
successful execution of cloud-init on each EC2 instance should be the working
Kubernetes cluster.

Buckets are created for each customer, then inside the buckets there are
folders for each cluster.

Each cluster contains the cloudconfig for master and worker instances.

For example, assuming that we have the following customers with clusters:

- first-customer
  - first-customers-cluster-1
  - first-customers-cluster-2
- second-customer
  - second-customers-cluster-1

we will have the following buckets with contents:

```
first-customer
|- first-customers-cluster-1
|  |- cloudconfig
|     |- master
|     |- worker
|- first-customers-cluster-2
|  |- cloudconfig
|     |- master
|     |- worker

second-customer
|- second-customers-cluster
|  |- cloudconfig
|     |- master
|     |- worker

```

## Secret

Here the plain passwords/tokens have to be inserted.
```
aws:
  accesskey:
    id: 'TODO'
    secret: 'TODO'
```

Here the base64 representation of the data structure above has to be inserted.
```
apiVersion: v1
kind: Secret
metadata:
  name: aws-operator-secret
  namespace: giantswarm
type: Opaque
data:
  secret.yml: 'TODO'
```

To create the secret manually do this.
```
kubectl create -f ./path/to/secret.yml
```

We also need a key to hold the SSH public key

```
apiVersion: v1
kind: Secret
metadata:
  name: aws-operator-ssh-key-secret
  namespace: giantswarm
type: Opaque
data:
  id_rsa.pub: 'TODO'
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/aws-operator/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

aws-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file
for details.

## Credit
- https://golang.org
- https://github.com/giantswarm/microkit
