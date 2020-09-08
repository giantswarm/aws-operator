# Policies

The `aws-operator` needs IAM permissions in order to properly manage tenant
clusters on AWS.

The recommended way of setting up the account for aws-operator is using [our
terraform modules].

If you prefer to do it manually see [our setup docs].

[our setup docs]: https://github.com/giantswarm/docs/blob/master/src/content/guides/prepare-aws-account-for-tenant-clusters/index.md#prepare-an-aws-account-to-run-giant-swarm-clusters
[our terraform modules]: https://github.com/giantswarm/giantswarm-aws-account-prerequisites
