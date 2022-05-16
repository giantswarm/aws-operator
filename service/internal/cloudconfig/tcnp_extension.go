package cloudconfig

import (
	"context"
	"encoding/base64"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/v6/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/v4/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v13/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudconfig/template"
	"github.com/giantswarm/aws-operator/service/internal/encrypter"
)

type TCNPExtension struct {
	awsConfigSpec g8sv1alpha1.AWSConfigSpec
	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	// See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	cc             *controllercontext.Context
	cluster        infrastructurev1alpha3.AWSCluster
	clusterCerts   []certs.File
	encrypter      encrypter.Interface
	encryptionKey  string
	externalSNAT   bool
	registryDomain string
}

func (e *TCNPExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	ctx := context.Background()

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
			Permissions: 0700,
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
			Permissions: 0700,
		},
		{
			// on worker NIC eth0  is used for machine and all other eth interfaces are for aws cni
			// add configuration for systemd-network to ignore aws cni interfaces
			AssetContent: fmt.Sprintf(template.NetworkdIgnoreAWSCNiInterfaces, "eth[1-9]*"),
			Path:         "/etc/systemd/network/00-ignore-aws-cni-interfaces.network",
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
			ctx = controllercontext.NewContext(ctx, *e.cc)

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
		AWSRegion:      key.Region(e.cluster),
		ExternalSNAT:   e.externalSNAT,
		IsChinaRegion:  key.IsChinaRegion(key.Region(e.cluster)),
		RegistryDomain: e.registryDomain,
	}

	for _, m := range filesMeta {
		c, err := k8scloudconfig.RenderFileAssetContent(m.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asset := k8scloudconfig.FileAsset{
			Metadata: m,
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

func (e *TCNPExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: template.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: template.PersistentVarLibDockerMount,
			Name:         "var-lib-docker.mount",
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
			AssetContent: template.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: template.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
		{
			AssetContent: template.EphemeralVarLibKubeletMount,
			Name:         "var-lib-kubelet.mount",
			Enabled:      true,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, m := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, e.awsConfigSpec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := k8scloudconfig.UnitAsset{
			Metadata: m,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *TCNPExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}
