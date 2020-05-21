package cloudconfig

import (
	"context"
	"encoding/base64"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v2/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudconfig/template"
	"github.com/giantswarm/aws-operator/service/internal/encrypter"
)

type TCCPNExtension struct {
	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	//     https://github.com/giantswarm/giantswarm/issues/4329.
	//
	cc               *controllercontext.Context
	cluster          infrastructurev1alpha2.AWSCluster
	clusterCerts     []certs.File
	encrypter        encrypter.Interface
	encryptionKey    string
	externalSNAT     bool
	masterID         int
	randomKeyTmplSet RandomKeyTmplSet
	registryDomain   string
}

func (e *TCCPNExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	ctx := controllercontext.NewContext(context.Background(), *e.cc)

	storageClass := template.InstanceStorageClassEncryptedContent

	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: template.DecryptTLSAssetsScript,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},
		{
			AssetContent: template.DecryptKeysAssetsScript,
			Path:         "/opt/bin/decrypt-keys-assets",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},

		{
			AssetContent: e.randomKeyTmplSet.APIServerEncryptionKey,
			Path:         "/etc/kubernetes/encryption/k8s-encryption-config.yaml.enc",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: template.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},
		// Add use-proxy-protocol to ingress-controller ConfigMap, this doesn't work
		// on KVM because of dependencies on hardware LB configuration.
		{
			AssetContent: template.IngressControllerConfigMap,
			Path:         "/srv/ingress-controller-cm.yml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		// NVME disks udev rules and script.
		// Workaround for https://github.com/coreos/bugs/issues/2399
		{
			AssetContent: template.NVMEUdevRule,
			Path:         "/etc/udev/rules.d/10-ebs-nvme-mapping.rules",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: template.NVMEUdevScript,
			Path:         "/opt/ebs-nvme-mapping",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0766,
		},
		{
			AssetContent: storageClass,
			Path:         "/srv/default-storage-class.yaml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: template.Etcd3ExtraConfig,
			Path:         "/etc/systemd/system/etcd3.d/10-require-attach-dep.conf",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: template.SystemdNetworkdEth1Network,
			Path:         "/etc/systemd/network/10-eth1.network",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: template.AwsCNIManifest,
			Path:         "/srv/aws-cni.yaml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
	}

	certsMeta := []k8scloudconfig.FileMetadata{}
	{
		for _, f := range e.clusterCerts {
			var encrypted string
			{
				e, err := e.encrypter.Encrypt(ctx, e.encryptionKey, string(f.Data))
				if err != nil {
					return nil, microerror.Mask(err)
				}
				encrypted = e
			}

			meta := k8scloudconfig.FileMetadata{
				AssetContent: encrypted,
				Path:         f.AbsolutePath + ".enc",
				Owner: k8scloudconfig.Owner{
					Group: k8scloudconfig.Group{
						Name: FileOwnerGroupName,
					},
					User: k8scloudconfig.User{
						Name: FileOwnerUserName,
					},
				},
				Permissions: 0700,
			}

			certsMeta = append(certsMeta, meta)
		}
	}

	var fileAssets []k8scloudconfig.FileAsset

	data := TemplateData{
		AWSRegion:            key.Region(e.cluster),
		ExternalSNAT:         e.externalSNAT,
		IsChinaRegion:        key.IsChinaRegion(key.Region(e.cluster)),
		MasterENIName:        key.ControlPlaneENIName(&e.cluster, e.masterID),
		MasterEtcdVolumeName: key.ControlPlaneVolumeName(&e.cluster, e.masterID),
		RegistryDomain:       e.registryDomain,
	}

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderFileAssetContent(fm.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asset := k8scloudconfig.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		fileAssets = append(fileAssets, asset)
	}

	for _, cm := range certsMeta {
		c := base64.StdEncoding.EncodeToString([]byte(cm.AssetContent))
		asset := k8scloudconfig.FileAsset{
			Metadata: cm,
			Content:  c,
		}

		fileAssets = append(fileAssets, asset)
	}

	return fileAssets, nil
}

func (e *TCCPNExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Create symlinks for nvme disks.
		// This service should be started only on first boot.
		{
			AssetContent: template.NVMEUdevTriggerUnit,
			Name:         "ebs-nvme-udev-trigger.service",
			Enabled:      true,
		},
		// Set bigger timeouts for NVME driver.
		// Workaround for https://github.com/coreos/bugs/issues/2484
		// TODO issue: https://github.com/giantswarm/giantswarm/issues/4255
		{
			AssetContent: template.NVMESetTimeoutsUnit,
			Name:         "nvme-set-timeouts.service",
			Enabled:      true,
		},
		{
			AssetContent: template.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: template.DecryptKeysAssetsService,
			Name:         "decrypt-keys-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: template.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: template.EphemeralVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enabled:      true,
		},
		// Attach etcd3 dependencies (EBS and ENI).
		{
			AssetContent: template.Etcd3AttachDepService,
			Name:         "etcd3-attach-dependencies.service",
			Enabled:      true,
		},
		// Automount etcd EBS volume.
		{
			AssetContent: template.AutomountEtcdVolume,
			Name:         "var-lib-etcd.automount",
			Enabled:      true,
		},
		// Mount etcd EBS volume.
		{
			AssetContent: template.MountEtcdVolumeAsgMasters,
			Name:         "var-lib-etcd.mount",
			Enabled:      false,
		},
		// Mount log EBS volume.
		{
			AssetContent: template.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	data := TemplateData{
		AWSRegion:            key.Region(e.cluster),
		ExternalSNAT:         e.externalSNAT,
		IsChinaRegion:        key.IsChinaRegion(key.Region(e.cluster)),
		MasterENIName:        key.ControlPlaneENIName(&e.cluster, e.masterID),
		MasterEtcdVolumeName: key.ControlPlaneVolumeName(&e.cluster, e.masterID),
		RegistryDomain:       e.registryDomain,
	}

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := k8scloudconfig.UnitAsset{
			Metadata: fm,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *TCCPNExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}
