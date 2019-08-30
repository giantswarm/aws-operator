package cloudconfig

import (
	"context"
	"encoding/base64"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_7_0"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter/vault"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/templates/cloudconfig"
)

// RandomKeyTmplSet holds a collection of rendered templates for random key
// encryption via KMS.
type RandomKeyTmplSet struct {
	APIServerEncryptionKey string
}

type MasterExtension struct {
	awsConfigSpec g8sv1alpha1.AWSConfigSpec
	baseExtension

	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	//	See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	ctlCtx *controllercontext.Context

	ClusterCerts     certs.Cluster
	RandomKeyTmplSet RandomKeyTmplSet
}

func (e *MasterExtension) Files(ctx context.Context) ([]k8scloudconfig.FileAsset, error) {
	// TODO Pass context to k8scloudconfig rendering functions.
	//
	//     https://github.com/giantswarm/giantswarm/issues/4329
	//
	var storageClass string
	_, ok := e.encrypter.(*vault.Encrypter)
	if ok {
		storageClass = cloudconfig.InstanceStorageClassContent
	} else {
		storageClass = cloudconfig.InstanceStorageClassEncryptedContent
	}

	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
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
			AssetContent: cloudconfig.DecryptKeysAssetsScript,
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
			AssetContent: e.RandomKeyTmplSet.APIServerEncryptionKey,
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
			AssetContent: cloudconfig.WaitDockerConf,
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
		{
			AssetContent: cloudconfig.VaultAWSAuthorizerScript,
			Path:         "/opt/bin/vault-aws-authorizer",
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
			AssetContent: cloudconfig.IngressControllerConfigMap,
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
			AssetContent: cloudconfig.NVMEUdevRule,
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
			AssetContent: cloudconfig.NVMEUdevScript,
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
	}

	certsMeta := []k8scloudconfig.FileMetadata{}
	{
		certFiles := certs.NewFilesClusterMaster(e.ClusterCerts)

		for _, f := range certFiles {
			// TODO We should just pass ctx to Files.
			//
			// 	See https://github.com/giantswarm/giantswarm/issues/4329.
			//
			ctx = controllercontext.NewContext(ctx, *e.ctlCtx)

			data, err := e.encrypt(ctx, f.Data)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			meta := k8scloudconfig.FileMetadata{
				AssetContent: string(data),
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

	data := e.templateData()

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

func (e *MasterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Create symlinks for nvme disks.
		// This service should be started only on first boot.
		{
			AssetContent: cloudconfig.NVMEUdevTriggerUnit,
			Name:         "ebs-nvme-udev-trigger.service",
			Enabled:      true,
		},
		// Set bigger timeouts for NVME driver.
		// Workaround for https://github.com/coreos/bugs/issues/2484
		// TODO issue: https://github.com/giantswarm/giantswarm/issues/4255
		{
			AssetContent: cloudconfig.NVMESetTimeoutsUnit,
			Name:         "nvme-set-timeouts.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsService,
			Name:         "decrypt-keys-assets.service",
			Enabled:      true,
		},

		{
			AssetContent: cloudconfig.VaultAWSAuthorizerService,
			Name:         "vault-aws-authorizer.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enabled:      true,
		},
		// Mount etcd EBS volume.
		{
			AssetContent: cloudconfig.MountEtcdVolume,
			Name:         "var-lib-etcd.mount",
			Enabled:      true,
		},
		// Mount log EBS volume.
		{
			AssetContent: cloudconfig.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, e.awsConfigSpec)
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
