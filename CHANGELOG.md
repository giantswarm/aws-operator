# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Fixed

- Added new Flatcar AMI identifiers.

## [9.3.6] - 2021-02-02

### Fixed

- Added CNI CIDR to internal ELB Security Group.

## [9.3.5] - 2020-12-08

### Changed

- Do not return NAT gateways in state `deleting` and `deleted` to avoid problems with recreating clusters with same ID.

## [9.3.4] - 2020-12-07

### Added

- Add vertical pod autoscaler support.
- Update `k8scloudconfig` version to `v9.3.0` to include change for cgroup for kubelet.

## [9.3.3] - 2020-12-02
### Changed

- Update `k8scloudconfig` version to `v9.2.0` to include change for kubelet pull QPS.

## [9.3.2] - 2020-11-26

### Changed

- Make it mandatory to configure alike instances via e.g. the installations repo.
- Fix naming and logs for `terminate-unhealthy-node` feature.

## [9.3.1-fix] - 2020-12-24

### Changed

- Remove explicit registry pull limits defaulting to less restrictive upstream settings.

## [9.3.1] - 2020-11-12

### Changed

- Update dependencies to next major versions.

### Fixed

- During a deletion of a cluster, ignore volumes that are mounted to an instance in a different cluster.

## [9.3.0] - 2020-11-09

### Added

- Annotation `alpha.aws.giantswarm.io/metadata-v2` to enable AWS Metadata API v2
- Annotation `alpha.aws.giantswarm.io/aws-subnet-size` to customize subnet size of Control Plane and Node Pools
- Annotation `alpha.aws.giantswarm.io/update-max-batch-size` to configure max batch size in ASG update policy on cluster or machine deployment CR.
- Annotation `alpha.aws.giantswarm.io/update-pause-time` to configure pause between batches in ASG update on cluster or machine deployment CR.

## [9.2.0] - 2020-11-03

### Added

- Annotation `alpha.giantswarm.io/aws-metadata-v2` to enable AWS Metadata API v2
- Add `terminate-unhealthy-node` feature to automatically terminate bad and
  unhealthy nodes in a Cluster.

### Fixed

- Fix dockerhub QPS by using paid user token for pulls.
- Remove dependency on `var-lib-etcd.automount` to avoid dependency cycle on
  new systemd.

## [9.1.3] - 2020-10-21

### Fixed

- Ignore error when missing APIServerPublicLoadBalancer CF Stack output to allow upgrade.

## [9.1.2] - 2020-10-15

### Added

- Add etcd client certificates for Prometheus.
- Add `--service.aws.hostaccesskey.role` flag.
- Add `api.<cluster ID>.k8s.<base domain>` and `*.<cluster ID>.k8s.<base domain>` records into CP internal hosted zone.

### Fixes

- Fix `vpc`/`route-table` lookups.

### Changed

- Access Control Plane AWS account using role assumption. This is to prepare
  running aws-operator inside a Tenant Cluster.
- Changed AWS CNI parameters to be more conservative with preallocated IPs while not hitting the AWS API too hard.

### Changed

- Update `k8scloudconfig` to `v8.0.3`.


## [9.1.1] - 2020-09-23

### Fixed

- Update flatcar AMI for China

## [9.1.0] - 2020-09-22

- Update AWS CNI manifests
- Disable Calico CNI binaries installation

## [9.0.1] - 2020-09-17

- Update flatcar releases

## [9.0.0] - 2020-09-15

### Added

- Emit Kubernetes events for tcnpf Cloudformation stack failures
- Emit Kubernetes events for tccpi and tccpf Cloudformation stack failures
- Add monitoring label
- Handle the case when there are both public and private hosted zones for CP
  base domain.
- Add Route Table lookup using tags, so `RouteTables` flag can be phased out in the future.


### Changed

- Update backward incompatible Kubernetes dependencies to v1.18.5.
- Remove migration code to ensure the Control Plane CRs for existing Node Pool clusters.

### Deprecated

- `RouteTables` flag will be deprecated.

### Fixed

- Don't panic when AWSControlPlane CR AZs are nil.
- Add suffix to Route Tables to get rid of naming collision.
- Fix image-pull-progress-deadline argument for tcnp nodes.

### Removed

- Remove etcd snapshot migration code.
- Remove unused `--service.aws.accesskey.id`, `--service.aws.accesskey.secret`
  and `--service.aws.accesskey.session` flags.
- Remove the prometheus collector and move it to the separate `aws-collector` project.

## [8.8.0] - 2020-08-14

- New version for a new kubernetes release.

## [8.7.6] - 2020-08-14

### Added

- Add release version tag for ec2 instances
- Update Cloudformation Stack when components version differ
- Emit Kubernetes events in case of change detection for tccp, tccpn and tcnp CF stacks

### Fixed

- Fix IAM policy on Tenant Clusters to manages IAM Role tags.
- Fixed passing custom pod CIDR to k8scloudconfig.

## [8.7.5] - 2020-07-30

### Changed

- Adjust number of host network pods on worker node for aws-cni

## [8.7.4] - 2020-07-29

### Fixed

- Adjust MAX_PODS for master and worker nodes to max IP's per ENI when using aws-cni

### Changed

- Use aws-cni version from the release.
- Use aws-cni image built based on https://github.com/giantswarm/aws-cni
- `k8scloudconfig` version updated to 7.0.4.

## [8.7.3] - 2020-07-15

### Fixed

- Fix regional switch in helm chart.

## [8.7.2] - 2020-07-14

### Added

- Add `--service.registry.mirrors` flag for setting registry mirror domains.
- Make registry domain & mirrors configurable based on region.

### Changed

- Replace `--service.registrydomain` with `--service.registry.domain`.
- Update `k8s-setup-network-env` image to `0.2.0`.

### Fixed

- Fix failing of ELB collector cache in case there is no ELB in AWS account


## [8.7.1] - 2020-07-08

### Added

- Add mapping between similar instance types `m4.16xlarge` and `m5.16xlarge`.
- Add `lifecycle` label to the `aws_operator_ec2_instance_status` metric to distinguish on-demand and spot.

### Changed

- Use `k8s-apiserver` image which includes CAs to enable OIDC.
- Use `0.1.0` tag for `aws-attach-etcd-dep` image.
- Use `0.1.0` tag for `k8s-setup-network-env` image.
- Use `0.1.1` tag for `k8s-api-healthz` image.

### Fixed

- Fix failing go template rendering of KMS encryption content.



## [8.7.0] 2020-06-19

### Added

- Add caching to the ELB collector.
- Add `keepforcrs` handler for more reliable CR cleanup.
- Add Control Plane labels to master nodes.
- Use the alpine 3.12 base Docker image

### Fixed

- Fix upgrade problems with pending volume snapshots.
- Fix cluster deletion issues in AWS using `DependsOn`.
- Fix calico-policy only metrics endpoint.
- Fix race condition in IPAM locking when lock already acquired.



## [8.6.1] 2020-05-21

### Added

- Add common labels to `aws-operator` pod.

### Fixed

- Fix collector panic.



## [8.6.0] 2020-05-21

### Added

- Enable ExternalSNAT to be configurable.

### Changed

- CI: Add optional pushing of WIP work to Aliyun registry.
- Remove static ip FOR ENI to avoid collision with internal API LB.
- Remove `--service.feature*` and `--service.test*` flags.

### Fixed

- Check Service Quota endpoint availability for the current AWS region
- Fix RBAC rules for Control Plane CR migration.



## [8.5.0] 2020-05-11

### Added

- Add common labels to our managed components.
- Disable profiling for Controller Manager and Scheduler.
- Add network policy.
- Move containerPort values from deployment to `values.yaml`.
- Enable per-cluster configuration of kube-proxy's `conntrackMaxPerCore` parameter.

### Changed

- Replace CoreOS with Flatcar.

### Fixed

- Fix cluster creation by preventing S3 Object upload race condition.



## [8.4.0] 2020-04-23

### Added

- Add mixed instance support for worker ASGs.

### Changed

- Improve cleanup of `DrainerConfig` CRs after node draining.
- Use release.Revision in Helm chart for Helm 3 support.



## [8.3.0] 2020-04-17

### Added

- Add Control Plane drainer controller.
- Add Dependabot configuration.
- Add VPC ID to AWSCluster CR status.
- Read CIDR from CR if available.

### Changed

- Drop CRD management to not ensure CRDs in operators anymore.

### Fixed

- Fix aws operator policy for latest node pools version.
- Make encryption key lookup graceful during cluster creation.



## [8.2.3] 2020-04-06

### Fixed

- Fix error handling when creating Tenant Cluster API clients.



## [8.2.2] - 2020-04-03

### Changed

- Switch from dep to Go modules.
- Use architect orb.
- Fix subnet allocation for Availability Zones.
- Switch to AWS CNI



## [8.2.1] - 2020-03-20

- Add PV limit per node. The limit is 20 PV per node.

### Added

- First release.


[9.3.6]: https://github.com/giantswarm/aws-operator/compare/v9.3.5...v9.3.6
[9.3.5]: https://github.com/giantswarm/aws-operator/compare/v9.3.4...v9.3.5
[9.3.4]: https://github.com/giantswarm/aws-operator/compare/v9.3.3...v9.3.4
[9.3.3]: https://github.com/giantswarm/aws-operator/compare/v9.3.2...v9.3.3
[9.3.2]: https://github.com/giantswarm/aws-operator/compare/v9.3.1...v9.3.2
[9.3.1-fix]: https://github.com/giantswarm/aws-operator/compare/v9.3.1...v9.3.1-fix
[9.3.1]: https://github.com/giantswarm/aws-operator/compare/v9.3.0...v9.3.1
[9.3.0]: https://github.com/giantswarm/aws-operator/compare/v9.2.0...v9.3.0
[9.2.0]: https://github.com/giantswarm/aws-operator/compare/v9.1.3...v9.2.0
[9.1.3]: https://github.com/giantswarm/aws-operator/compare/v9.1.2...v9.1.3
[9.1.2]: https://github.com/giantswarm/aws-operator/compare/v9.1.1...v9.1.2
[9.1.1]: https://github.com/giantswarm/aws-operator/compare/v9.1.0...v9.1.1
[9.1.0]: https://github.com/giantswarm/aws-operator/compare/v9.0.1...v9.1.0
[9.0.1]: https://github.com/giantswarm/aws-operator/compare/v9.0.0...v9.0.1
[9.0.0]: https://github.com/giantswarm/aws-operator/compare/v8.8.0...v9.0.0
[8.8.0]: https://github.com/giantswarm/aws-operator/compare/v8.7.6...v8.8.0
[8.7.6]: https://github.com/giantswarm/aws-operator/compare/v8.7.5...v8.7.6
[8.7.5]: https://github.com/giantswarm/aws-operator/compare/v8.7.4...v8.7.5
[8.7.4]: https://github.com/giantswarm/aws-operator/compare/v8.7.3...v8.7.4
[8.7.3]: https://github.com/giantswarm/aws-operator/compare/v8.7.2...v8.7.3
[8.7.2]: https://github.com/giantswarm/aws-operator/compare/v8.7.1...v8.7.2
[8.7.1]: https://github.com/giantswarm/aws-operator/compare/v8.7.0...v8.7.1
[8.7.0]: https://github.com/giantswarm/aws-operator/compare/v8.6.1...v8.7.0
[8.6.1]: https://github.com/giantswarm/aws-operator/compare/v8.6.0...v8.6.1
[8.6.0]: https://github.com/giantswarm/aws-operator/compare/v8.5.0...v8.6.0
[8.5.0]: https://github.com/giantswarm/aws-operator/compare/v8.4.0...v8.5.0
[8.4.0]: https://github.com/giantswarm/aws-operator/compare/v8.3.0...v8.4.0
[8.3.0]: https://github.com/giantswarm/aws-operator/compare/v8.2.3...v8.3.0
[8.2.3]: https://github.com/giantswarm/aws-operator/compare/v8.2.2...v8.2.3
[8.2.2]: https://github.com/giantswarm/aws-operator/compare/v8.2.1...v8.2.2

[8.2.1]: https://github.com/giantswarm/aws-operator/releases/tag/v8.2.1
