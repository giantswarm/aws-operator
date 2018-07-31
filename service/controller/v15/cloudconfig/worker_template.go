package cloudconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_5_0"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v15/templates/cloudconfig"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(ctx context.Context, customObject v1alpha1.AWSConfig, certs legacy.CompactTLSAssets) (string, error) {
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
			certs:         certs,
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
	certs legacy.CompactTLSAssets
}

func (e *WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        "root:root",
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.WorkerCrt,
			Path:         "/etc/kubernetes/ssl/worker-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.WorkerCA,
			Path:         "/etc/kubernetes/ssl/worker-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.WorkerKey,
			Path:         "/etc/kubernetes/ssl/worker-key.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     GzipBase64Encoding,
			Permissions:  0700,
		},
		{
			AssetContent: cloudconfig.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        "root:root",
			Permissions:  0700,
		},
	}

	var newFiles []k8scloudconfig.FileAsset

	for _, m := range filesMeta {
		data := e.templateData()
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := k8scloudconfig.FileAsset{
			Metadata: m,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *WorkerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
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
