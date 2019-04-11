# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

The latest version is considered WIP and it is a subject of change. All other
versions are frozen. To freeze current version all files are copied to a new
version directory, and  then changes are introduced.

## [v4.3.0] WIP

## [v4.2.0]

### Changed
- Fix race condition issue with systemd units.

### Removed

- Remove `UsePrivilegeSeparation` option from sshd configuration.

## [v4.1.2]
### Changed
- Pin calico-kube-controllers to master.
- Fix calico-node felix severity log level.
- Enable `serviceaccount` controller in calico-kube-controller.
- Remove 'staticPodPath' from worker kubelet configuration.

## [v4.1.1]
### Changed
- Update kubernetes to 1.13.4 CVE-2019-1002100

## [v4.1.0]
### Changed
- Intall calicoctl, crictl and configure etcctl tooling in masters.
- Update kubernetes to 1.13.3.
- Update etcd to 3.3.12.
- Update calico to 3.5.1.
- Add fine-grained Audit Policy

## [v4.0.1]
### Changed
- Update kubernetes to 1.12.6 CVE-2019-1002100

## [v3.8.0] WIP
- Update kubernetes to 1.12.6 CVE-2019-1002100

## [v4.0.0]

### Changed
- Switched from cloudinit to ignition.
- Double the inotify watches.
- Switch kube-proxy from `iptables` to `ipvs`.
- Add worker node labels.
- Increase timeouts for etcd defragmentaion.

### Removed

- Ingress Controller and CoreDNS manifests. Now migrated to chart-operator.
- Removed nodename_file_optional from calico configmap.

## [v3.7.5]
- Update kubernetes to 1.12.6 CVE-2019-1002100

## [v3.7.4]

### Changed
- Double the inotify watches.

### Removed
- Removed nodename_file_optional from calico configmap.

## [v3.7.3]

### Changed
- update kubernetes to 1.12.3 (CVE-2018-1002105)

## [v3.6.4]

### Changed
- Update `libreadline` version

## [v3.6.3]
- update kubernetes to 1.11.5 (CVE-2018-1002105)

### Changed
- update kubernetes to 1.10.11 (CVE-2018-1002105)

## [v3.5.3]

### Changed
- Update `libreadline` version

## [v3.5.2]

### Changed

## [v3.7.2]

### Changed
- Remove the old master from the k8s api before upgrading calico (k8s-addons)
- Wait until etcd DNS is resolvable before upgrading calico. Networking pods crashlooping isn't fun!

## [v3.7.1]

### Changed
- The pod priority class for calico got lost. We found it again!
- kube-proxy is now installed before calico during cluster creation and upgrades.

## [v3.7.0]

### Changed
- Updated Kubernetes to 1.12.2
- Updated etcd to 3.3.9
- Kubernetes and etcd images are now held in one place
- Updated audit policy version
- Moved audit policy out of static pod path
- Updated rbac resources to v1
- Remove static pod path from worker nodes
- Remove readonly port from kubelet
- Add DBUS socket and ClusterCIDR to kube-proxy

## [v3.6.2]

### Changed
- Updated Calico to 3.2.3
- Updated Calico manifest with resource limits to get QoS policy guaranteed.
- Enabled admission plugins: DefaultTolerationSeconds, MutatingAdmissionWebhook, ValidatingAdmissionWebhook.

## [v3.6.1]

### Changed
- Use patched GiantSwarm build of Kubernetes (`hyperkube:v1.11.1-cec4fb8023db783fbf26fb056bf6c76abfcd96cf-giantswarm`).

## [v3.6.0]

### Added
- Added template flag for removing CoreDNS resources (will be managed by
chart-operator).

### Changed
- Updated Kubernetes (hyperkube) to version 1.11.1.
- Updated Calico to version 3.2.0.

### Removed


## [v3.5.1]


## [v3.5.0]

### Changed
- Disabled HSTS headers in nginx-ingress-controller.
- Added separate parameter for disabling the Ingress Controller service manifest.

### Removed


## [v3.4.0]

### Added
- Added SSO public key into ssh trusted CA.
- Added RBAC rules for node-operator.
- Added RBAC rules for prometheus.
- Enabled monitoring for ingress controller metrics.
- Parameterize image registry domain.
- Set "worker-processes" to 4 for nginx-ingress-controller.
- Added `--feature-gates=CustomResourceSubresources=true` to enable status subresources for CRDs.

### Changed
- Add memory eviction thresholds for kubelet in order to preserve node in case of heavy load.
- Updated etcd version to 3.3.8

### Removed


## [v3.3.4]

### Changed
- Added parameter for disabling Ingress Controller related components.
- Increased start timeout for k8s-kubelet.service.

### Removed


## [v3.3.3]

### Changed

- Remove Nginx version from `Server` header in Ingress Controller
- Updated hyperkube to version 1.10.4.

### Removed


## [v3.3.2]

### Changed
- Updated hyperkube to version 1.10.2 due to regression in 1.10.3 with configmaps.

### Removed
- Removed node-exporter related components (will be managed by chart-operator).

## [v3.3.1]

### Changed
- Changed some remaining images to be pulled from Giant Swarm's registry.
- Updated Alpine sidecar for Ingress Controller to version 3.7.
- Fixed mkfs.xfs for containerized kubelet.
- Updated Kubernetes (hyperkube) to version 1.10.3.

### Removed


## [v3.3.0]

### Changed
- Updated hyperkube to version 1.10.2.

### Removed
- Removed kube-state-metrics related components (will be managed by
chart-operator).


## [v3.2.6]

### Changed
- Changed node-exporter to have named ports.
- Added RBAC rules for configmaps, secrets and hpa for kube-state-metrics.
- Synced privileged PSP with upstream (adding all added capabilities and seccomp profiles)
- Downgraded hyperkube to version 1.9.5.

### Removed


## [v3.2.5]

### Changed
- Updated kube-state-metrics to version 1.3.1.
- Updated hyperkube to version 1.10.1.
- Changed kubelet bind mount mode from "shared" to "rshared".
- Disabled etcd3-defragmentation service in favor systemd timer.
- Added /lib/modules mount for kubelet.
- Updated CoreDNS to 1.1.1.
- Fixed node-exporter running in container by adjusting host mounts.
- Updated Calico to 3.0.5.
- Updated Etcd to 3.3.3.
- Added trusted certificate CNs to aggregation API allowed names.
- Disabled SSL passthrough in nginx-ingress-controller.
- Removed explicit hostname labeling for kubelet.

### Removed
- Removed docker flag "--disable-legacy-registry".
- Removed calico-ipip-pinger.


## [v3.2.4]

### Changed
- Masked systemd-networkd-wait-online unit.
- Makes injection of Kubernetes encryption key configurable.
- Enabled volume resizing feature.



## [v3.2.3]

### Changed
- Updated Kubernetes with version 1.9.5.
- Updated nginx-ingress-controller to version 0.12.0.

### Removed
- Removed hard limits from core kubernetes components.



## [v3.2.2]

### Removed
- Removed set-ownership-etcd-data-dir.service.



## [v3.2.1]

### Added
- Added priority classes core-components, critical-pods and important pods.
- Added Guaranteed QoS for api/scheduler/controller-manager pods by assigning resources limits.

### Changed
- Enabled aggregation layer in Kubernetes API server.
- Ordered Kubernetes cluster components scheduling process by assigning PriorityClass to pods.

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


[v4.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_4_1_0
[v4.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_4_0_0
[v3.7.4]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_7_4
[v3.7.3]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_7_3
[v3.6.4]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_6_4
[v3.6.3]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_6_3
[v3.5.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_5_2
[v3.7.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_7_2
[v3.7.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_7_1
[v3.7.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_7_0
[v3.6.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_6_2
[v3.6.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_6_1
[v3.6.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_6_0
[v3.5.3]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_5_3
[v3.5.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_5_0
[v3.4.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_4_0
[v3.3.4]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_3_4
[v3.3.3]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_3_3
[v3.3.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_3_2
[v3.3.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_3_1
[v3.3.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_3_0
[v3.2.6]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_6
[v3.2.5]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_5
[v3.2.4]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_4
[v3.2.3]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_3
[v3.2.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_2
[v3.2.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_2_1
[v3.1.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_1_1
[v3.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_1_0
[v3.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_3_0_0
[v2.0.2]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_2_0_2
[v2.0.1]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_2_0_1
[v2.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v2
[v1.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v1_1
[v1.0.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v1
[v0.1.0]: https://github.com/giantswarm/k8scloudconfig/commits/master/v_0_1_0
