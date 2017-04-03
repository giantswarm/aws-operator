[![CircleCI](https://circleci.com/gh/giantswarm/aws-operator.svg?&style=shield&circle-token=8f0fe6ad08c090afa36c35ba5d926ac6ffe797e8)](https://circleci.com/gh/giantswarm/aws-operator)

# aws-operator

The aws-operator handles Kubernetes clusters running on a Kubernetes cluster inside of AWS.

## Prerequisites

## Getting Project

Download the latest release: https://github.com/giantswarm/aws-operator/releases/latest

Clone the git repository: https://github.com/giantswarm/aws-operator.git

Download the latest docker image from here: https://hub.docker.com/r/giantswarm/aws-operator/

### How to build

This project provides a Makefile, so you can build it by typing:

```
make
```

If you prefer, you may also build it using the standard `go build` command, like:

```
go build github.com/giantswarm/awstpr
```

## Running aws-operator

After building the project, you will have a `aws-operator` binary.

The operator needs some Kubernetes secrets to be present. The secrets contain
the TLS assets (CAs, keys, certificates) for the various components of the
cluster.

An easy way to create these secrets for development is running:


```
kubectl create -f examples/secrets.yml
```

Afterwards, you can run:

```
./aws-operator daemon --aws.accesskey.id <aws_acces_key_id> --aws.accesskey.secret <aws_access_key_secret> --aws.region <aws_region>
```

In the future, we are going to use aws-operator as a Kubernetes pod and that would be the standard
way of usage.

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/aws-operator/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the contribution workflow as well as reporting bugs.

## License

aws-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Credit
- https://golang.org
- https://github.com/giantswarm/microkit
