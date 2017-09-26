package cloudconfig

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/microerror"
)

func v_0_1_0WorkerTemplate(customObject awstpr.CustomObject, certs certificatetpr.CompactTLSAssets) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v_0_1_0WorkerExtension{
			certs: certs,
		}
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.WorkerTemplate, params)
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

type v_0_1_0WorkerExtension struct {
	certs certificatetpr.CompactTLSAssets
}

func (e *v_0_1_0WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
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
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.WorkerCA,
			Path:         "/etc/kubernetes/ssl/worker-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.WorkerKey,
			Path:         "/etc/kubernetes/ssl/worker-key.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  0700,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     k8scloudconfig.GzipBase64,
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
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, nil)
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

func (e *v_0_1_0WorkerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
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
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, nil)
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

func (e *v_0_1_0WorkerExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	newSections := []k8scloudconfig.VerbatimSection{
		{
			Name:    "storageclass",
			Content: instanceStorageClassTemplate,
		},
	}

	return newSections
}
