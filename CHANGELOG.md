# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

*Note: This changelog only tracks changes in the `legacy-1-15` branch.*

## [Unreleased]

### Fixed

- Fix conntrack configuration structure for `kube-proxy`.
- Fix regression when applying `kube-proxy` manifest during cluster upgrades.

### Changed

- Replace CoreOS with Flatcar Container Linux.

## [5.5.2] - 2020-05-06

### Fixed

- Update error package for the double-quote issue.

### Changed

- Added back functionality to push docker images to aliyun.

## [5.5.1] - 2020-03-30

### Added

- Enable per-cluster configuration of kube-proxy's `conntrackMaxPerCore` parameter.

### Changed

- Streamline image templating for core components for quicker and easier releases in the future.
- Retrieve component versions from `releases`.


## [5.5.0] - 2019-11-01

### Added

- First release as a flattened operator.

### Changed

- Update to Kubernetes 1.15.5.


[Unreleased]: https://github.com/giantswarm/aws-operator/compare/v5.5.2...legacy-1-15
[5.5.2]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.2
[5.5.1]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.1
[5.5.0]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.0
