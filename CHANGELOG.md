# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

*Note: This changelog only tracks changes in the `legacy` branch.*

## [Unreleased]

## [5.7.5] - 2020-08-28

### Added

- Updated AMI mapping to add Flatcar Container Linux 2512.3.0 AMIs.

### Changed

- Removed memory and CPU limits from calico-kube-controllers.

## [5.7.4] - 2020-08-19

- Check subnets from node pools clusters when collecting allocated subnets.

## [5.7.2] - 2020-07-29

- Add support for latest Container Linux AMIs.

## [5.7.1] - 2020-05-07

### Fixed

- Correct conntrack configuration structure for kube-proxy.

## [5.7.0] - 2020-04-30

### Added

- Support for Flatcar Linux.
- Enable per-cluster configuration of kube-proxy's `conntrackMaxPerCore` parameter.

### Changed

- Streamline image templating for core components for quicker and easier releases in the future.
- Retrieve component versions from `releases`.
- Update helm chart with modern labels, configuration, and templating.
- Switch from dep to Go modules.
- Use release.Revision in Helm chart for Helm 3 support.
- Only replace nodes when ignition changes.

### Removed

- Management of AWSConfig CRD.

## [5.6.0] - 2020-01-29

### Changed

- Update to Kubernetes 1.16.3.


## [5.5.0] - 2019-11-01

### Added

- First release as a flattened operator.

### Changed

- Update to Kubernetes 1.15.5.


[Unreleased]: https://github.com/giantswarm/aws-operator/compare/v5.7.5...legacy
[5.7.5]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.5
[5.7.4]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.4
[5.7.3]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.3
[5.7.2]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.2
[5.7.1]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.1
[5.7.0]: https://github.com/giantswarm/aws-operator/releases/tag/v5.7.0
[5.6.0]: https://github.com/giantswarm/aws-operator/releases/tag/v5.6.0
[5.5.0]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.0
