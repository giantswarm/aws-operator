# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] 2020-02-12

### Added

- Add JSON function.

### Changed

- Error should be defined as a value instead of pointer.
- Maskf takes only Error type.
- Use built-in errors package instead of juju/errgo.
- Print error stacks in JSON format instead of custom errgo format.
- Switch from dep to Go modules.

### Removed

- Drop Error.String method.
- Drop New function.
- Drop Newf function.
- Drop Stack function in favour of JSON function.

## [0.1.0] 2020-02-03

### Added

- First release.

[Unreleased]: https://github.com/giantswarm/microerror/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/giantswarm/microerror/releases/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/microerror/releases/tag/v0.1.0
