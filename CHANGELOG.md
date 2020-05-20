# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

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



[Unreleased]: https://github.com/giantswarm/aws-operator/compare/v8.5.0...HEAD

[8.5.0]: https://github.com/giantswarm/aws-operator/compare/v8.4.0...v8.5.0
[8.4.0]: https://github.com/giantswarm/aws-operator/compare/v8.3.0...v8.4.0
[8.3.0]: https://github.com/giantswarm/aws-operator/compare/v8.2.3...v8.3.0
[8.2.3]: https://github.com/giantswarm/aws-operator/compare/v8.2.2...v8.2.3
[8.2.2]: https://github.com/giantswarm/aws-operator/compare/v8.2.1...v8.2.2

[8.2.1]: https://github.com/giantswarm/aws-operator/releases/tag/v8.2.1
