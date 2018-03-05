# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

The latest version is considered WIP and it is a subject of change. All other
versions are frozen. To freeze current version all files are copied to a new
version directory, and then changes are introduced.

## [v3.2.1] WIP

## [v3.1.1]

### Added
- Added calico-ipip-pinger.

### Changed
- Change etcd data path to /var/lib/etcd.
- Fix `StartLimitIntervalSec` parameter location in `etcd3` systemd unit.
- Add `feature-gates` flag in api server enabling `ExpandPersistentVolumes` feature.
- Updated calico to 3.0.2.
- Updated etcd to 3.3.1.
- Tune kubelet flags for protecting key units (kubelet and container runtime) from workload overloads.
- Updated nginx-ingress-controller to 0.11.0.
- Updated coredns to 1.0.6.

## [v3.1.0]

### Changed
- Systemd units for Kubernetes components (api-server, scheduler and controller-manager)
  replaced with self-hosted pods.

## [v3.0.0]

### Added
- Add encryption config template for API etcd data encryption experimental
  feature.
- Add tests for template evaluations.
- Allow adding extra manifests.
- Allow extract Hyperkube options.
- Allow setting custom K8s API address for master nodes.
- Allow setting etcd port.
- Add node-exporter.
- Add kube-state-metrics.

### Changed
- Unify CloudConfig struct construction.
- Update calico to 3.0.1.
- Update hyperkube to v1.9.2.
- Update nginx-ingress-controller to 0.10.2.
- Use vanilla (previously coreos) hyperkube image.
- kube-dns replaced with CoreDNS 1.0.5.
- Fix Kubernetes API audit log.

### Removed
- Remove calico-ipip-pinger.
- Remove calico-node-controller.

## [v2.0.2]

### Added
- Add fix for scaled workers to ensure they have a kube-proxy.

## [v2.0.1]

### Changed
- Fix audit logging.

## [v2.0.0]

### Added
- Disable API etcd data encryption experimental feature.

### Changed
- Updated calico to 2.6.5.

### Removed
- Removed calico-ipip-pinger.
- Removed calico-node-controller.

## [v1.1.0]

### Added
- Use Cluster type from https://github.com/giantswarm/apiextensions.

## [v1.0.0]

### Removed
- Disable API etcd data encryption experimental feature.

## [v0.1.0]

[v3.1.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_1_1
[v3.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_1_0
[v3.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_0_0
[v2.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v2
[v1.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v1_1
[v1.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v1
[v0.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_0_1_0
