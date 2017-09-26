package deployment

import (
	"fmt"
	"path/filepath"

	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func newMasterDeployments(customObject kvmtpr.CustomObject) ([]*extensionsv1.Deployment, error) {
	var deployments []*extensionsv1.Deployment

	privileged := true
	replicas := int32(1)

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		capabilities := customObject.Spec.KVM.Masters[i]

		storageType := key.StorageType(customObject)

		// During migration, some TPOs do not have storage type set.
		// This specifies a default, until all TPOs have the correct storage type set.
		// tl;dr - this shouldn't be here. If all TPOs have storageType, remove it.
		if storageType == "" {
			storageType = "hostPath"
		}

		var etcdVolume apiv1.Volume
		if storageType == "hostPath" {
			etcdVolume = apiv1.Volume{
				Name: "etcd-data",
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: key.MasterHostPathVolumeDir(key.ClusterID(customObject), key.VMNumber(i)),
					},
				},
			}
		} else if storageType == "persistentVolume" {
			etcdVolume = apiv1.Volume{
				Name: "etcd-data",
				VolumeSource: apiv1.VolumeSource{
					PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
						ClaimName: key.EtcdPVCName(key.ClusterID(customObject), key.VMNumber(i)),
					},
				},
			}
		} else {
			return nil, microerror.Maskf(wrongTypeError, "unknown storageType: '%s'", key.StorageType(customObject))
		}
		deployment := &extensionsv1.Deployment{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "extensions/v1beta",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.DeploymentName(key.MasterID, masterNode.ID),
				Labels: map[string]string{
					"cluster":  key.ClusterID(customObject),
					"customer": key.ClusterCustomer(customObject),
					"app":      key.MasterID,
					"node":     masterNode.ID,
				},
			},
			Spec: extensionsv1.DeploymentSpec{
				Strategy: extensionsv1.DeploymentStrategy{
					Type: extensionsv1.RecreateDeploymentStrategyType,
				},
				Replicas: &replicas,
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: apismetav1.ObjectMeta{
						GenerateName: key.MasterID,
						Labels: map[string]string{
							"app":      key.MasterID,
							"cluster":  key.ClusterID(customObject),
							"customer": key.ClusterCustomer(customObject),
							"node":     masterNode.ID,
						},
						Annotations: map[string]string{},
					},
					Spec: apiv1.PodSpec{
						Affinity:    newMasterPodAfinity(customObject),
						HostNetwork: true,
						NodeSelector: map[string]string{
							"role": key.MasterID,
						},
						Volumes: []apiv1.Volume{
							{
								Name: "cloud-config",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: key.ConfigMapName(customObject, masterNode, key.MasterID),
										},
									},
								},
							},
							etcdVolume,
							{
								Name: "images",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: "/home/core/images/",
									},
								},
							},
							{
								Name: "rootfs",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: filepath.Join("/home/core/vms", key.ClusterID(customObject), masterNode.ID),
									},
								},
							},
						},
						Containers: []apiv1.Container{
							{
								Name:            "k8s-endpoint-updater",
								Image:           customObject.Spec.KVM.EndpointUpdater.Docker.Image,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								Command: []string{
									"/opt/k8s-endpoint-updater",
									"update",
									"--provider.bridge.name=" + key.NetworkBridgeName(customObject),
									"--service.kubernetes.cluster.namespace=" + key.ClusterNamespace(customObject),
									"--service.kubernetes.cluster.service=" + key.MasterID,
									"--service.kubernetes.inCluster=true",
								},
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
							},
							{
								Name:            "k8s-kvm",
								Image:           customObject.Spec.KVM.K8sKVM.Docker.Image,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Args: []string{
									key.MasterID,
								},
								Env: []apiv1.EnvVar{
									{
										Name:  "CORES",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "DISK",
										Value: fmt.Sprintf("%.0fG", capabilities.Disk),
									},
									{
										Name: "HOSTNAME",
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
												APIVersion: "v1",
												FieldPath:  "metadata.name",
											},
										},
									},
									{
										Name:  "NETWORK_BRIDGE_NAME",
										Value: key.NetworkBridgeName(customObject),
									},
									{
										Name: "MEMORY",
										// TODO provide memory like disk as float64 and format here.
										Value: capabilities.Memory,
									},
									{
										Name:  "ROLE",
										Value: key.MasterID,
									},
									{
										Name:  "CLOUD_CONFIG_PATH",
										Value: "/cloudconfig/user_data",
									},
								},
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      "cloud-config",
										MountPath: "/cloudconfig/",
									},
									{
										Name:      "etcd-data",
										MountPath: "/etc/kubernetes/data/etcd/",
									},
									{
										Name:      "images",
										MountPath: "/usr/code/images/",
									},
									{
										Name:      "rootfs",
										MountPath: "/usr/code/rootfs/",
									},
								},
							},
						},
					},
				},
			},
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
