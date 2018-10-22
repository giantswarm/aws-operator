package cloudconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_6_2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v18/templates/cloudconfig"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(ctx context.Context, customObject v1alpha1.AWSConfig, clusterCerts certs.Cluster) (string, error) {
	var err error

	encryptionKey, err := c.encrypter.EncryptionKey(ctx, customObject)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var params k8scloudconfig.Params
	{
		be := baseExtension{
			customObject:  customObject,
			encrypter:     c.encrypter,
			encryptionKey: encryptionKey,
		}

		params.Cluster = customObject.Spec.Cluster
		params.Extension = &WorkerExtension{
			baseExtension: be,

			ClusterCerts: clusterCerts,
		}
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = c.k8sKubeletExtraArgs
		params.RegistryDomain = c.registryDomain
		params.SSOPublicKey = c.SSOPublicKey
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = k8scloudconfig.WorkerTemplate

		newCloudConfig, err = k8scloudconfig.NewCloudConfig(cloudConfigConfig)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return newCloudConfig.Base64(), nil
}

type WorkerExtension struct {
	baseExtension

	ClusterCerts certs.Cluster
}

func (e *WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	// TODO https://github.com/giantswarm/giantswarm/issues/4329
	ctx := context.TODO()

	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        "root:root",
			Permissions:  0700,
		},
		{
			AssetContent: cloudconfig.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        "root:root",
			Permissions:  0700,
		},
	}

	{
		certFiles := certs.NewFilesClusterWorker(e.ClusterCerts)

		for _, f := range certFiles {
			data, err := e.encryptAndGzip(ctx, f.Data)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			meta := k8scloudconfig.FileMetadata{
				AssetContent: string(data),
				Path:         f.AbsolutePath,
				Owner:        FileOwner,
				Permissions:  0700,
			}

			filesMeta = append(filesMeta, meta)
		}
	}

	var fileAssets []k8scloudconfig.FileAsset

	for _, m := range filesMeta {
		data := e.templateData()
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asset := k8scloudconfig.FileAsset{
			Metadata: m,
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
			Enable:       false,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.WorkerFormatVarLibDockerService,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.PersistentVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
		// Set bigger timeouts for NVME driver.
		// Workaround for https://github.com/coreos/bugs/issues/2484
		// TODO issue: https://github.com/giantswarm/giantswarm/issues/4255
		{
			AssetContent: cloudconfig.NVMESetTimeoutsUnit,
			Name:         "nvme-set-timeouts.service",
			Enable:       true,
			Command:      "start",
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, m := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, e.customObject.Spec)
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
	newSections := []k8scloudconfig.VerbatimSection{
		{
			Name:    "storageclass",
			Content: cloudconfig.InstanceStorageClass,
		},
	}

	return newSections
}
