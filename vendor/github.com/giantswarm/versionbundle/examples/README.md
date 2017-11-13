# examples

This folder contains examples showing how version bundles look like. The
[kubernetes-operator.json](kubernetes-operator.json) shows the version bundles
returned by the `kubernetes-operator`. It defines which versions it provides in
which combinations with respect to its expected dependencies. The
[cloud-config-operator.json](cloud-config-operator.json) shows the same for
another microservice called `cloud-config-operator`. The
[aggregation.json](aggregation.json) shows how the version bundles of the two
mentioned operators are aggregated to reflect the combined version bundles as
they can be used together. These aggregations are then used as releases.
