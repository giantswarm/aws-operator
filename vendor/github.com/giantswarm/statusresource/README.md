[![CircleCI](https://circleci.com/gh/giantswarm/statusresource.svg?&style=shield&circle-token=b91250c237a9800cbeeacd1a54e8bb1def458355)](https://circleci.com/gh/giantswarm/statusresource)

# statusresource

Package statusresource implements primitives for CR status management within
Giant Swarm Kubernetes guest clusters.

## Statuses

Statuses are written under `.status.cluster.conditions` in the CR,
in the following format :

```
    - lastTransitionTime: "2019-07-05T07:42:24.076879904Z"
      status: "True"
      type: Created
```

Allowed status type are defined in [apiextensions](https://godoc.org/github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1) as `StatusClusterType*`.

Decisions for status transition are made based on number and version of nodes.
