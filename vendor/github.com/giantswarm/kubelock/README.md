[![GoDoc](https://godoc.org/github.com/giantswarm/kubelock?status.svg)](http://godoc.org/github.com/giantswarm/kubelock) [![CircleCI](https://circleci.com/gh/giantswarm/kubelock.svg?style=shield)](https://circleci.com/gh/giantswarm/kubelock)

# kubelock

Package kubelock provides functionality to create distributed locks on
arbitrary kubernetes resources. It is heavily inspired by [pulcy/kube-lock] but
uses [client-go] library and its dynamic client.

# Usage

At Giant Swarm we run multiple instances of the same operators in different
versions. Some actions performed by operators are not concurrent. Good example
is IP range allocation for a newly created cluster. Each lock created by
kubelock has a name and an owner. A custom name allows to create multiple locks
on the same Kubernetes resource. The owner string usually contains the version
of the operator acquiring the lock. That way the operator can know if the lock
was acquired by itself or other operator version.

[client-go]: https://github.com/kubernetes/client-go
[pulcy/kube-lock]: https://github.com/pulcy/kube-lock
