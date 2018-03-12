[![CircleCI](https://circleci.com/gh/giantswarm/aws-operator.svg?&style=shield&circle-token=8f0fe6ad08c090afa36c35ba5d926ac6ffe797e8)](https://circleci.com/gh/giantswarm/aws-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/aws-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/aws-operator)


# aws-operator

The aws-operator manages Kubernetes clusters running on AWS.

## Getting Project

Download the latest release:
https://github.com/giantswarm/aws-operator/releases/latest

Clone the git repository: https://github.com/giantswarm/aws-operator.git

Download the latest docker image from here:
https://quay.io/repository/giantswarm/aws-operator


### How to build

Build the standard way.

```
go build github.com/giantswarm/aws-operator
```

## Running aws-operator

See [this guide][examples-local].

[examples-local]: https://github.com/giantswarm/aws-operator/blob/master/examples

## Architecture

The operator uses our [operatorkit](1) framework. It manages an `awsconfig`
CRD using a generated client stored in our [apiextensions](2) repo. Releases
are versioned using [version bundles](3).

The operator provisions guest Kubernetes clusters running on AWS. It runs in a
host Kubernetes cluster also running on AWS.

[1]:https://github.com/giantswarm/operatorkit
[2]:https://github.com/giantswarm/apiextensions
[3]:https://github.com/giantswarm/versionbundle

### CloudFormation

The guest Kubernetes clusters are provisioned using [AWS CloudFormation](4). The
resources are split between 3 CloudFormation stacks.

* guest-main manages the guest cluster resources.
* host-setup manages an IAM role used for VPC peering.
* host-main manages network routes for the VPC peering connection.

The host cluster may run in a separate AWS account. If so resources are created
in both the host and guest AWS accounts.

[4]:https://aws.amazon.com/cloudformation

### Other AWS Resources

As well as the CloudFormation stacks we also provision a KMS key and S3 bucket
per cluster. This is to upload cloudconfigs for the cluster nodes. The
cloudconfigs contain TLS certificates which are encrypted using the KMS key.

### Kubernetes Resources

The operator also creates a Kubernetes namespace per guest cluster with a
service and endpoints. These are used by the host cluster to access the guest
cluster.

### Certificates

Authentication for the cluster components and end-users uses TLS certificates.
These are provisioned using [Hashicorp Vault](5) and are managed by our
[cert-operator](6).

[5]:https://www.vaultproject.io/
[6]:https://github.com/giantswarm/cert-operator

## Secret

Here the AWS IAM credentials have to be inserted.
```
service:
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
