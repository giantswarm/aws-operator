[![GoDoc](https://godoc.org/github.com/giantswarm/helmclient?status.svg)](http://godoc.org/github.com/giantswarm/helmclient) [![CircleCI](https://circleci.com/gh/giantswarm/helmclient.svg?&style=shield)](https://circleci.com/gh/giantswarm/helmclient)

# helmclient

Package helmclient implements [Helm] related primitives to work against helm
releases. Currently supports Helm 2 and connects to the Tiller gRPC API using
a port forwarding connection.

## Interface

See `helmclient.Interface` in [spec.go] for supported methods.

## Related libraries

[k8sportforward] is used to establish a port forwarding conection with Tiller. 

## License

helmclient is under the Apache 2.0 license. See the [LICENSE](LICENSE) file
for details.

[Helm]: https://github.com/helm/helm
[k8sportforward]: https://github.com/giantswarm/k8sportforward
[spec.go]: https://github.com/giantswarm/helmclient/blob/master/spec.go
