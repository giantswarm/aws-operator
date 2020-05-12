package cloudconfig

import (
	"context"
	"encoding/base64"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	cloudconfig "github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig/template"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCNPConfig struct {
	Config Config
}

type TCNP struct {
	config Config
}

func NewTCNP(config TCNPConfig) (*TCNP, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &TCNP{
		config: config.Config,
	}

	return t, nil
}

func (t *TCNP) Render(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, images k8scloudconfig.Images, labels string) ([]byte, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var kubeletExtraArgs []string
	{
		if t.config.PodInfraContainerImage != "" {
			kubeletExtraArgs = append(kubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", t.config.PodInfraContainerImage))
		}

		kubeletExtraArgs = append(kubeletExtraArgs, t.config.KubeletExtraArgs...)
	}

	// Allow the actual externalSNAT to be set by the CR.
	var externalSNAT bool
	if key.ExternalSNAT(cr) == nil {
		externalSNAT = t.config.ExternalSNAT
	} else {
		externalSNAT = *key.ExternalSNAT(cr)
	}

	var params k8scloudconfig.Params
	{
		// Default registry, kubernetes, etcd images etcd.
		// Required for proper rending of the templates.
		params = k8scloudconfig.DefaultParams()

		params.Cluster = cmaClusterToG8sConfig(t.config, cr, labels).Cluster
		params.Extension = &WorkerExtension{
			awsConfigSpec: cmaClusterToG8sConfig(t.config, cr, labels),
			baseExtension: baseExtension{
				cluster:        cr,
				encrypter:      t.config.Encrypter,
				encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
				externalSNAT:   externalSNAT,
				registryDomain: t.config.RegistryDomain,
			},
			cc:           cc,
			clusterCerts: clusterCerts,
		}
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.Images = images
		params.EnableAWSCNI = true
		params.SSOPublicKey = t.config.SSOPublicKey

		ignitionPath := k8scloudconfig.GetIgnitionPath(t.config.IgnitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templateBody []byte
	{
		c := k8scloudconfig.CloudConfigConfig{
			Params:   params,
			Template: k8scloudconfig.WorkerTemplate,
		}

		cloudConfig, err := k8scloudconfig.NewCloudConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = cloudConfig.ExecuteTemplate()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		templateBody = []byte(cloudConfig.String())
	}

	return templateBody, nil
}

type WorkerExtension struct {
	awsConfigSpec g8sv1alpha1.AWSConfigSpec
	baseExtension
	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	// See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	cc           *controllercontext.Context
	clusterCerts certs.Cluster
}

func (e *WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	ctx := context.TODO()

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
			Permissions: 0700,
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
			Permissions: 0700,
		},
	}

	certsMeta := []k8scloudconfig.FileMetadata{}
	{
		certFiles := certs.NewFilesClusterWorker(e.clusterCerts)

		for _, f := range certFiles {
			ctx = controllercontext.NewContext(ctx, *e.cc)

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

	data := e.templateDataTCNP()

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

func (e *WorkerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.PersistentVarLibDockerMount,
			Name:         "var-lib-docker.mount",
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
			AssetContent: cloudconfig.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibKubeletMount,
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

func (e *WorkerExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	newSections := []k8scloudconfig.VerbatimSection{}

	return newSections
}
