# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [16.1.1] - 2024-04-02

### Fixed

- Bump k8scc to fix issues with IMDS v2.

## [16.1.0] - 2024-03-04

### Changed

- Bump k8scc to avoid running etcd defrag on all masters at the same time.

### Fixed

- Handle karpenter nodes in node-termination-handler.

## [16.0.0] - 2024-01-16

### Changed

- Bump k8scc to v18 to enable k8s 1.25 support.

## [15.0.0] - 2023-12-13

### Fixed

- Bump k8scc to fix calculation of max pods per node when in ENI mode.

### Changed

- [Breaking change] Removed code that allowed switching from AWS-CNI to Cilium. Releases using this AWS-operator can't be upgraded to from v18 releases.
- Configure `gsoci.azurecr.io` as the default container image registry.

## [14.24.1] - 2023-12-07

### Changed

- Bump etcd-cluster-migrator to v1.2.0

## [14.24.0] - 2023-11-20

### Added

- Add `global.podSecurityStandards.enforced` value for PSS migration.
- Emit event when an unhealthy node is terminated.
- Bump `badnodedetector` to be able to use `node-problem-detector` app for unhealthy node termination.
- Add a additional IAM permission for `cluster-autoscaler`.

### Changed

- Bump k8scc to disable PSPs in preparation for switch to PSS.
- Disable cluster autoscaler during rollouts of node pool ASGs.

## [14.23.0] - 2023-10-04

### Fixed

- Cleanup `kube-proxy` VPA after switching to Cilium.
- Bump k8scc to enable max pod calculations when cilium is in ENI IPAM mode.

## [14.22.0] - 2023-09-11

### Changed

- Get AMI data from helm value rather than from hardcoded string in the code.

## [14.21.0] - 2023-09-01

## [14.20.0] - 2023-08-29

### Added

- Allow newer flatcar releases for node pools as provided by AWS release.
- Add `sigs.k8s.io/cluster-api-provider-aws/role` tag to all subnets as preparation for migration to CAPI.

### Changed

- Unmanage interfaces for CNI eth[1-9] on workers eth[2-9] on masters
- [cilium eni mode] Only run aws-node, calico and kube-proxy on old nodes during migration to cilium.

## [14.19.2] - 2023-08-03

### Fixed

- Update vulnerable net package.

## [14.19.1] - 2023-08-03

### Fixed

- Fix rule names of PolicyException.

### Changed

- Update IAM policy for AWS LoadBalancer Controller.

## [14.19.0] - 2023-07-14

### Added

- Enable auditd.

## [14.18.0] - 2023-07-13

### Added

- Add support for customizing `controller-manager` `terminated-pod-gc-threshold` value through annotation `controllermanager.giantswarm.io/terminated-pod-gc-threshold`
- Add Service Monitor.

### Changed

- Check if all nodes are rolled before deleting AWS CNI resources when upgrading from v18 to v19.
- Change Route53 Trust Policy to allow other applications to use the role.

### Fixed

- Change AWS LB Controller Trust Policy for the new S3 bucket.
- Added pss exceptions for volumes.

## [14.17.1] - 2023-05-11

### Added

- Add toleration for new control-plane taint.

### Fixed

- Ensure `net.ipv4.conf.eth0.rp_filter` is set to `2` if aws-CNI is used.
- Make `routes-fixer` script compatible with alpine.

## [14.17.0] - 2023-05-05

### Changed

- Disable ETCD compaction request from apiserver.

## [14.16.0] - 2023-05-03

### Changed

- Do not delete aws-cni subnets when switching to cilium.

## [14.15.0] - 2023-04-25

### Fixed

- Allow to enable ACLs for a S3 buckets.

## [14.14.0] - 2023-04-19

### Added

- Added ami IDs for flatcar `3510.2.0`.

## [14.13.0] - 2023-04-18

### Fixed

- Use `alpine` as image for aws-cni's `routes-fixer`. 

### Changed

- Allow externalDNS role to be assumed by any SA containing "external-dns" to allow multiple app deployments.

## [14.12.1] - 2023-04-05

### Added

- Added ami IDs for flatcar `3374.2.4` and `3374.2.5`.

## [14.12.0] - 2023-04-04

### Changed

- Set ENV for nftables in `aws-cni`.

## [14.11.0] - 2023-04-04

### Changed

- Improved etcd resiliency and allow customization of `--quota-backend-bytes`.

## [14.10.0] - 2023-03-16

### Changed

- When creating a cluster, create the master ASGs in parallel.

## [14.9.0] - 2023-03-07

### Changed

- Bump k8s-api-healthz image to 0.2.0.

### Fixed

- Don't mark master instance as unhealthy if local etcd instance is unresponsive but the whole etcd cluster is also down.
- Don't mark master instance as unhealthy if local API server instance is unresponsive but the whole API server is also down.

## [14.8.0] - 2023-03-02

### Changed

- Adjust the tccpn stack heartbeat to improve cluster upgrades.

## [14.7.1] - 2023-02-03

### Changed

- Switch container registry in China

## [14.7.0] - 2023-02-02

### Added

- Label node pool nodes with `cgroups.giantswarm.io/version` to indicate which cgroup version they are running.

## [14.6.0] - 2023-01-30

### Fixed

- Adjust ALBController IAM role name.

### Changed

- Add AMIs for flatcar versions 3374.2.1, 3374.2.2 and 3374.2.3.

## [14.5.0] - 2023-01-26

### Added

- Add `ALB Controller` IAM role.

### Changed

- Update k8scloudconfig to allow setting custom kernel parameters in the 'net.*' namespace.
- Remove IP limit when prefix delegation is enabled. IP limit will be 110 for nodes with Prefix Delegation.

### Added

- Allow disk size configuration of logging volume. New default value is 15Gb.
- Allow different values for docker and containerd volume.

### Fixed

- Fix Docker rate limit for pulling images.

## [14.4.0] - 2023-01-13

### Changed

- Bump k8scc to 15.4.0.

## [14.3.0] - 2022-11-29

### Added

- Add flatcar 3374.2.0.

### Changed

- Bump k8scc to 15.3.0.

## [14.2.0] - 2022-11-24

### Changed

- Bump k8scc to 15.2.0.

## [14.1.0] - 2022-11-16

### Changed

- Use custom KMS key for encryption on your Amazon EBS volumes.
- Enable IRSA by default in release v19.0.0.
- Bump k8scc to 15.1.1.
- Added EFS policy to the ec2 instance role to allow to use the EFS driver out of the box
- Add both the cloudfront domain and alias domain in route53manager role policy.

### Fixed

- Allow rolling nodes when there is a change in the AWSMachineDeployment even when CF stack was never updated before.
- Quickly delete DrainerConfigs during cluster or machine deployment deletion to speedup cluster deletion process.
- Fix disabling of kube-proxy in v19+.

## [14.0.0] - 2022-10-11

### Added

- Add AMI reference for flatcar 3227.2.2.
- Lifecycle hook for launching master instances in HA mode.

### Changed

- Bump k8scc to 15.0.0.
- Disable kube-proxy on release v19 and newer.
- Allow master node to change the autoscaling healthcheck.

### Fixed

- Fix node draining logic during node termination.

## [13.2.4] - 2022-10-27

### Changed

- Add old cloudfront domain name as service-account-issuer when domain alias is enabled in IRSA.

## [13.2.3] - 2022-10-24

### Changed

- Avoid duplicate `--service-account-signing-key-file` flag being set for API server.

## [13.2.2] - 2022-10-21

### Fixed

- Add cluster API endpoint as sts audience.

## [13.2.1] - 2022-08-31

### Fixed

- Bump k8scc to support cgroups v1 on containerd.

## [13.2.0] - 2022-08-29

## [13.1.0] - 2022-08-25

### Changed

- Enable Cilium or AWS-CNI conditionally based on the release number.
- Disable external cloud controller manager because of upstream bug affecting 1.23 release.
- Bump `k8scc` to enable authn and authz on `scheduler` and `controller-manager`.

## [13.0.0] - 2022-08-17

### Changed

- Use Cloudfront Domain for IRSA for non-China regions.
- Ensure `aws-node` daemonset does not schedule on upgraded nodes.
- Ensure `aws-node` daemonset has `AWS_VPC_K8S_CNI_EXCLUDE_SNAT_CIDRS` env var set to the cilium cidr during migration to cilium.
- Cleanup `aws-node` resources after a successful migration.
- Cleanup `calico` resources after a successful migration.
- Use `cilium.giantswarm.io/pod-cidr` annotation as Cilium Pod CIDR.
- Add Flatcar `3227.2.1` AMI.
- Bump `k8scloudconfig` to support newer flatcar.
- Set EC2's `HttpPutResponseHopLimit` flag to 2.

### Removed

- Remove creation of cilium app config.

## [13.0.0-alpha2] - 2022-07-27

### Changed

- Bump k8scc to fix apiserver's flags and make metrics-server to work.

## [13.0.0-alpha1] - 2022-07-25

### Added

- Added new flatcar 3227.2.0 image release.

### Changed

- Revert applying external cloud controller manager as a static pod.
- Disable calico and aws-cni.
- Create configmap to configure cilium app.
- Enable controller-manager's allocate-cidrs flag.

## [12.1.0] - 2022-07-18

### Added

- Containerd EBS Volume.

### Fixed

- Fix `crictl.yaml` on worker nodes.

## [12.0.0] - 2022-07-14

### Added

- Use external cloud controller manager for AWS.

### Changed

- Mount containerd socket instead of dockershim one to `aws-node` pods.

## [11.16.0] - 2022-07-04

### Added

- Added new flatcar 3139.2.3 image release.

### Changed

- Tighten pod and container security contexts for PSS restricted policies.
- Bump `k8scc` to enable `auditd` monitoring for `execve` syscalls.

## [11.15.0] - 2022-06-21

### Changed

- Set default upgrade batch to 10% from 33%
- Set default pause time to 10 minutes

## [11.14.1] - 2022-06-15

### Fixed

- Fix principal ARN for Route53 trusted entity.

### Changed

- Remove `imagePullSecrets`

## [11.14.0] - 2022-06-14

### Added 

- Added new flatcar 3139.2.2 image release.

## [11.13.0] - 2022-06-09

### Changed

- Bumped k8scc to latest version to fix localhost node name problem.

## [11.12.0] - 2022-05-25

### Added

- Extend permission policy of IAM role `Route53Manager-Role` for IRSA.

### Changed

- Bump `k8scc` to use `systemd` cgroup driver on masters and cgroups v2 worker nodes.
- Bump `aws-attach-etcd-dep` to 0.4.0.

## [11.11.0] - 2022-05-16

### Changed

- Update dependencies.

## [11.10.0] - 2022-05-11

### Added

- Set optionally the `kubernetes.io/role/internal-elb` tag to machine deployment subnets.

## [11.9.3] - 2022-05-02

### Fixed

- Set `AWS_VPC_K8S_CNI_RANDOMIZESNAT` to `prng` when SNAT is enabled.

## [11.9.2] - 2022-04-20

### Fixed

- Issuer S3 endpoint for IRSA.

## [11.9.1] - 2022-04-20

### Fixed

- AWS Region Endpoint for IRSA.

## [11.9.0] - 2022-04-20

### Added

- Add `POD_SECURITY_GROUP_ENFORCING_MODE` to `aws-node` Daemonset.

## [11.8.0] - 2022-04-19

### Added

- Added separate service account flag for IRSA.

## [11.7.0] - 2022-04-12

### Added

- Added latest flatcar images.

## [11.6.0] - 2022-04-12

### Changed

- Ignore S3 bucket deletion for audit logs.

## [11.5.0] - 2022-04-05

### Removed

- Remove tag `kubernetes.io/role/internal-elb` from machine deployment subnets.

## [11.4.0] - 2022-04-04

### Changed

- Bumped k8scc to 13.4.0 to enable VPA for kube-proxy.

## [11.3.0] - 2022-04-01

### Changed

- Bumped k8scc to 13.3.0 to disable VPA for kube-proxy and fix chicken-egg problem.

## [11.2.0] - 2022-04-01

### Changed

- Bumped k8scc to 13.2.0 to enable VPA for kube-proxy.

## [11.1.0] - 2022-03-31

### Added

- Add annotation to ASG to make cluster-autoscaler work when scaling from zero replicas.

## [11.0.0] - 2022-03-29

### Changed

- Update CAPI dependencies.

## [10.19.0] - 2022-03-21

### Addded

- Add latest flatcar AMIs.

### Changed

- Allow resource limits/requests to be passed as values.
- Switch `gp2` to `gp3` volumes.
- Allow etcd volume IOPS and Throughput to be set.

## [10.18.0] - 2022-03-04

### Added

- Add support for IAM Roles for Service Accounts feature.

## [10.17.0] - 2022-02-16

### Changed

- Bumped `k8scloudconfig` to disable `rpc-statd` service.

## [10.16.0] - 2022-02-14

### Added

- New flatcar releases.

## [10.15.1] - 2022-02-02

### Fixed

- Autoselect region ARN for ebs snapshots.

## [10.15.0] - 2022-02-01

### Added

- Add support for feature that enables forcing cgroups v1 for Flatcar version `3033.2.0` and above.

### Changed

- Bump `k8scloudconfig` version to `v11.0.1`.

## [10.14.0] - 2022-01-27

### Changed

- Changes to EncryptionConfig in order to fully work with `encryption=provider-operator`.

## [10.13.0] - 2022-01-19

### Changed

- Bump `k8scloudconfig` to latest release to support Calico 3.21.

## [10.12.0] - 2022-01-18

### Changed

- Max pods setting per for new EC2 instances.
- Bump `etcd-cluster-migrator` version to `v1.1.0`.

## [10.11.0] - 2022-01-05

### Added

- Add AMI for `af-south-1` region.
- IAM permission for EBS snapshots.

### Fixed

- Adjusted aws-cni manifests to support version 1.10.1.

## [10.10.1] - 2021-11-29

### Fixed

- Setting `kubernetes.io/replace/internal-elb` tag on private subnet TCNP stack.

## [10.10.0] - 2021-11-23

### Added

- Adding latest flatcar images.
- Introduce AWS CNI Prefix delegation.

### Changed

- Use k8smetadata for annotations.

## [10.9.1] - 2021-09-29

### Added

- Add cloud tags propagation to S3 buckets.

### Changed

-  Update `aws-attach-etcd-dep` image version to `0.2.0` to include bugfixes.

## [10.9.0] - 2021-09-28

### Added

- Add provider tags to the AWS CNI ENIs.
- Add configuration for `systemd-networkd` to ignore network interfaces used for AWS CNI.
- Add changes to run properly on Flatcar 2905 and newer.

### Changed

- Upgrade `k8scloudconfig` which is required for k8s 1.21.

## [10.8.0] - 2021-08-30

### Changed

- Introducing `v1alpha3` CR's.
- Update Flatcar AMI's to the latest stable releases.

## [10.7.1] - 2021-08-17

## [10.7.0] - 2021-08-11

### Added

- Add security settings to S3 bucket to comply with aws policies `s3-bucket-public-read-prohibited,s3-bucket-ssl-requests-only,s3-bucket-public-write-prohibited,s3-bucket-server-side-encryption-enabled,s3-bucket-logging-enabled`, `aws-operator` will need additonal permissions `s3:PutBucketPublicAccessBlock` and `s3:PutBucketPolicy`.

## [10.6.1] - 2021-07-01

## Changed

- Upgrade `k8scloudconfig` to v10.8.1 which includes a change to better determine if memory eviction thresholds are crossed.

## [10.6.0] - 2021-06-29

### Added

- S3 vpc endpoint to AWS CNI subnet.

### Changed

- Update Flatcar AMI's to the latest stable releases.

## [10.5.0] - 2021-05-27

### Added

- Enabled EBS CSI migration.

### Removed

- Removed default storage-class annotation, EBS CSI driver is taking over.

## [10.4.0] - 2021-05-25

### Changed

- Avoid TCCPN stack failure by checking if a control-plane tag exists before adding it.
- Look up cloud tags in all namespaces
- Find certs in all namespaces
- Enable `terminate unhealthy node` feature by default.
- Add node termination counter per cluster metric.

## [10.3.0] - 2021-05-13

### Fixed

- Updated OperatorKit to v4.3.1 for Kubernetes 1.20 support.
- Cancel update loop if source or target release is not found.
- Updated IPAM library to avoid IP conflicts.

### Added

- Clean up VPC peerings from a cluster VPC when is cluster deleted.
- Clean up Application and Network loadbalancers created by Kubernetes when cluster is deleted.
- Add new flatcar AMIs.

### Changed

- Fix issues with etcd initial cluster resolving into ELB and causing errors.
- Update `k8scloudconfig` to version `v10.5.0` to support kubernetes `v1.20`.
- Use `networkctl reload` for managing networking to avoid bug in `systemd`.

## [10.2.0] - 2021-02-08

### Added

- Allow incoming NFS traffic on node pools for EFS.

## [10.1.0] - 2021-02-03

### Added

- Add support for tagging AWS resources, managed by the operator, based on the custom resource labels.

### Changed

- Use values generated by `config-controller` to deploy `aws-operator` instead of catalog values.
- Use `giantswarm/config` versions matching `v1.x.x` major.
- Start updating `tcnp` CF stack only when `tccpn` CF stack is already updated. This ensure that master nodes are updated before worker nodes.

## [10.0.0] - 2021-01-22

### Added

- Add `cleanupiamroles` resource for detaching third party policies from our IAM
  roles.
- Update `k8scloudconfig` version to `v10.0.0` to include change for Kubernetes 1.19.
- Allow configuration of `MINIMUM_IP_TARGET` and `WARM_IP_TARGET` for AWS CNI via annotations on `AWSCluster`

### Changed

- Include Account ID in the s3bucket for access logs. It is a breaking change, that will put access logs to a new s3 bucket.
- Change AWS CNI and AWS CNI k8s plugin log verbosity to `INFO`.
- Change AWS CNI log file to `stdout`.
- Add retry logic for decrypt units to avoid flapping.

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



[Unreleased]: https://github.com/giantswarm/aws-operator/compare/v16.1.1...HEAD
[16.1.1]: https://github.com/giantswarm/aws-operator/compare/v16.1.0...v16.1.1
[16.1.0]: https://github.com/giantswarm/aws-operator/compare/v16.0.0...v16.1.0
[16.0.0]: https://github.com/giantswarm/aws-operator/compare/v15.0.0...v16.0.0
[15.0.0]: https://github.com/giantswarm/aws-operator/compare/v14.24.1...v15.0.0
[14.24.1]: https://github.com/giantswarm/aws-operator/compare/v14.24.0...v14.24.1
[14.24.0]: https://github.com/giantswarm/aws-operator/compare/v14.23.0...v14.24.0
[14.23.0]: https://github.com/giantswarm/aws-operator/compare/v14.22.0...v14.23.0
[14.22.0]: https://github.com/giantswarm/aws-operator/compare/v14.21.0...v14.22.0
[14.21.0]: https://github.com/giantswarm/aws-operator/compare/v14.20.0...v14.21.0
[14.20.0]: https://github.com/giantswarm/aws-operator/compare/v14.19.2...v14.20.0
[14.19.2]: https://github.com/giantswarm/aws-operator/compare/v14.19.1...v14.19.2
[14.19.1]: https://github.com/giantswarm/aws-operator/compare/v14.19.0...v14.19.1
[14.19.0]: https://github.com/giantswarm/aws-operator/compare/v14.18.0...v14.19.0
[14.18.0]: https://github.com/giantswarm/aws-operator/compare/v14.17.1...v14.18.0
[14.17.1]: https://github.com/giantswarm/aws-operator/compare/v14.17.0...v14.17.1
[14.17.0]: https://github.com/giantswarm/aws-operator/compare/v14.16.0...v14.17.0
[14.16.0]: https://github.com/giantswarm/aws-operator/compare/v14.15.0...v14.16.0
[14.15.0]: https://github.com/giantswarm/aws-operator/compare/v14.14.0...v14.15.0
[14.14.0]: https://github.com/giantswarm/aws-operator/compare/v14.13.0...v14.14.0
[14.13.0]: https://github.com/giantswarm/aws-operator/compare/v14.12.1...v14.13.0
[14.12.1]: https://github.com/giantswarm/aws-operator/compare/v14.12.0...v14.12.1
[14.12.0]: https://github.com/giantswarm/aws-operator/compare/v14.11.0...v14.12.0
[14.11.0]: https://github.com/giantswarm/aws-operator/compare/v14.10.0...v14.11.0
[14.10.0]: https://github.com/giantswarm/aws-operator/compare/v14.9.0...v14.10.0
[14.9.0]: https://github.com/giantswarm/aws-operator/compare/v14.8.0...v14.9.0
[14.8.0]: https://github.com/giantswarm/aws-operator/compare/v14.7.1...v14.8.0
[14.7.1]: https://github.com/giantswarm/aws-operator/compare/v14.7.0...v14.7.1
[14.7.0]: https://github.com/giantswarm/aws-operator/compare/v14.6.0...v14.7.0
[14.6.0]: https://github.com/giantswarm/aws-operator/compare/v14.5.0...v14.6.0
[14.5.0]: https://github.com/giantswarm/aws-operator/compare/v14.4.0...v14.5.0
[14.4.0]: https://github.com/giantswarm/aws-operator/compare/v14.3.0...v14.4.0
[14.3.0]: https://github.com/giantswarm/aws-operator/compare/v14.2.0...v14.3.0
[14.2.0]: https://github.com/giantswarm/aws-operator/compare/v14.1.0...v14.2.0
[14.1.0]: https://github.com/giantswarm/aws-operator/compare/v14.0.0...v14.1.0
[14.0.0]: https://github.com/giantswarm/aws-operator/compare/v13.2.4...v14.0.0
[13.2.4]: https://github.com/giantswarm/aws-operator/compare/v13.2.3...v13.2.4
[13.2.3]: https://github.com/giantswarm/aws-operator/compare/v13.2.2...v13.2.3
[13.2.2]: https://github.com/giantswarm/aws-operator/compare/v13.2.1...v13.2.2
[13.2.1]: https://github.com/giantswarm/aws-operator/compare/v13.2.0...v13.2.1
[13.2.0]: https://github.com/giantswarm/aws-operator/compare/v13.1.0...v13.2.0
[13.1.0]: https://github.com/giantswarm/aws-operator/compare/v13.0.0...v13.1.0
[13.0.0]: https://github.com/giantswarm/aws-operator/compare/v13.0.0-alpha2...v13.0.0
[13.0.0-alpha2]: https://github.com/giantswarm/aws-operator/compare/v13.0.0-alpha1...v13.0.0-alpha2
[13.0.0-alpha1]: https://github.com/giantswarm/aws-operator/compare/v12.1.0...v13.0.0-alpha1
[12.1.0]: https://github.com/giantswarm/aws-operator/compare/v12.0.0...v12.1.0
[12.0.0]: https://github.com/giantswarm/aws-operator/compare/v11.16.0...v12.0.0
[11.16.0]: https://github.com/giantswarm/aws-operator/compare/v11.15.0...v11.16.0
[11.15.0]: https://github.com/giantswarm/aws-operator/compare/v11.14.1...v11.15.0
[11.14.1]: https://github.com/giantswarm/aws-operator/compare/v11.14.0...v11.14.1
[11.14.0]: https://github.com/giantswarm/aws-operator/compare/v11.13.0...v11.14.0
[11.13.0]: https://github.com/giantswarm/aws-operator/compare/v11.12.0...v11.13.0
[11.12.0]: https://github.com/giantswarm/aws-operator/compare/v11.11.0...v11.12.0
[11.11.0]: https://github.com/giantswarm/aws-operator/compare/v11.10.0...v11.11.0
[11.10.0]: https://github.com/giantswarm/aws-operator/compare/v11.9.3...v11.10.0
[11.9.3]: https://github.com/giantswarm/aws-operator/compare/v11.9.2...v11.9.3
[11.9.2]: https://github.com/giantswarm/aws-operator/compare/v11.9.1...v11.9.2
[11.9.1]: https://github.com/giantswarm/aws-operator/compare/v11.9.0...v11.9.1
[11.9.0]: https://github.com/giantswarm/aws-operator/compare/v11.8.0...v11.9.0
[11.8.0]: https://github.com/giantswarm/aws-operator/compare/v11.7.0...v11.8.0
[11.7.0]: https://github.com/giantswarm/aws-operator/compare/v11.6.0...v11.7.0
[11.6.0]: https://github.com/giantswarm/aws-operator/compare/v11.5.0...v11.6.0
[11.5.0]: https://github.com/giantswarm/aws-operator/compare/v11.4.0...v11.5.0
[11.4.0]: https://github.com/giantswarm/aws-operator/compare/v11.3.0...v11.4.0
[11.3.0]: https://github.com/giantswarm/aws-operator/compare/v11.2.0...v11.3.0
[11.2.0]: https://github.com/giantswarm/aws-operator/compare/v11.1.0...v11.2.0
[11.1.0]: https://github.com/giantswarm/aws-operator/compare/v11.0.0...v11.1.0
[11.0.0]: https://github.com/giantswarm/aws-operator/compare/v10.19.0...v11.0.0
[10.19.0]: https://github.com/giantswarm/aws-operator/compare/v10.18.0...v10.19.0
[10.18.0]: https://github.com/giantswarm/aws-operator/compare/v10.17.0...v10.18.0
[10.17.0]: https://github.com/giantswarm/aws-operator/compare/v10.16.0...v10.17.0
[10.16.0]: https://github.com/giantswarm/aws-operator/compare/v10.15.1...v10.16.0
[10.15.1]: https://github.com/giantswarm/aws-operator/compare/v10.15.0...v10.15.1
[10.15.0]: https://github.com/giantswarm/aws-operator/compare/v10.14.0...v10.15.0
[10.14.0]: https://github.com/giantswarm/aws-operator/compare/v10.13.0...v10.14.0
[10.13.0]: https://github.com/giantswarm/aws-operator/compare/v10.12.0...v10.13.0
[10.12.0]: https://github.com/giantswarm/aws-operator/compare/v10.11.0...v10.12.0
[10.11.0]: https://github.com/giantswarm/aws-operator/compare/v10.10.1...v10.11.0
[10.10.1]: https://github.com/giantswarm/aws-operator/compare/v10.10.0...v10.10.1
[10.10.0]: https://github.com/giantswarm/aws-operator/compare/v10.9.1...v10.10.0
[10.9.1]: https://github.com/giantswarm/aws-operator/compare/v10.9.0...v10.9.1
[10.9.0]: https://github.com/giantswarm/aws-operator/compare/v10.8.0...v10.9.0
[10.8.0]: https://github.com/giantswarm/aws-operator/compare/v10.7.1...v10.8.0
[10.7.1]: https://github.com/giantswarm/aws-operator/compare/v10.7.0...v10.7.1
[10.7.0]: https://github.com/giantswarm/aws-operator/compare/v10.6.1...v10.7.0
[10.6.1]: https://github.com/giantswarm/aws-operator/compare/v10.6.0...v10.6.1
[10.6.0]: https://github.com/giantswarm/aws-operator/compare/v10.5.0...v10.6.0
[10.5.0]: https://github.com/giantswarm/aws-operator/compare/v10.4.0...v10.5.0
[10.4.0]: https://github.com/giantswarm/aws-operator/compare/v10.3.0...v10.4.0
[10.3.0]: https://github.com/giantswarm/aws-operator/compare/v10.2.0...v10.3.0
[10.2.0]: https://github.com/giantswarm/aws-operator/compare/v10.1.0...v10.2.0
[10.1.0]: https://github.com/giantswarm/aws-operator/compare/v10.0.0...v10.1.0
[10.0.0]: https://github.com/giantswarm/aws-operator/compare/v9.3.5...v10.0.0
[9.3.5]: https://github.com/giantswarm/aws-operator/compare/v9.3.4...v9.3.5
[9.3.4]: https://github.com/giantswarm/aws-operator/compare/v9.3.3...v9.3.4
[9.3.3]: https://github.com/giantswarm/aws-operator/compare/v9.3.2...v9.3.3
[9.3.2]: https://github.com/giantswarm/aws-operator/compare/v9.3.1...v9.3.2
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
