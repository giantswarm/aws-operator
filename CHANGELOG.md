# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

*Note: This changelog only tracks changes in the `legacy-1-15` branch.*

## [Unreleased]

## [5.5.1] - 2020-03-30

### Added

- Enabled per-cluster configuration of kube-proxy's `conntrackMaxPerCore` parameter.

### Changed

- Streamlined image templating for core components for quicker and easier releases in the future.
- Retrieve component versions from `releases`.


## [5.5.0] - 2019-11-01

### Added

- First release as a flattened operator.

### Changed

- Updated to Kubernetes 1.15.5.


[Unreleased]: https://github.com/giantswarm/aws-operator/compare/v5.5.1...legacy-1-15
[5.5.1]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.1
[5.5.0]: https://github.com/giantswarm/aws-operator/releases/tag/v5.5.0
