package cloudconfigv3

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_0_0"
	"github.com/giantswarm/microerror"
)

const (
	// WorkerCloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	WorkerCloudConfigVersion = "v_3_0_0"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(customObject v1alpha1.AWSConfig, certs certificatetpr.CompactTLSAssets) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &WorkerExtension{
			certs:        certs,
			customObject: customObject,
		}
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
	certs        certificatetpr.CompactTLSAssets
	customObject v1alpha1.AWSConfig
}

func (e *WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: decryptTLSAssetsScriptTemplate,
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
			AssetContent: waitDockerConfTemplate,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        "root:root",
			Permissions:  0700,
		},
	}

	var newFiles []k8scloudconfig.FileAsset

	for _, m := range filesMeta {
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, e.customObject.Spec)
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
			AssetContent: decryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: workerFormatVarLibDockerServiceTemplate,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: persistentVarLibDockerMountTemplate,
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
			Content: instanceStorageClassTemplate,
		},
	}

	return newSections
}
