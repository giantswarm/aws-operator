package cloudconfig

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

func v_0_1_0MasterTemplate(customObject awstpr.CustomObject, certs certificatetpr.CompactTLSAssets, keys randomkeytpr.CompactRandomKeyAssets) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v_0_1_0MasterExtension{
			certs:        certs,
			customObject: customObject,
			keys:         keys,
		}
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.MasterTemplate, params)
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

type v_0_1_0MasterExtension struct {
	certs        certificatetpr.CompactTLSAssets
	customObject awstpr.CustomObject
	keys         randomkeytpr.CompactRandomKeyAssets
}

func (e *v_0_1_0MasterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: decryptTLSAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerCrt,
			Path:         "/etc/kubernetes/ssl/apiserver-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerCA,
			Path:         "/etc/kubernetes/ssl/apiserver-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerKey,
			Path:         "/etc/kubernetes/ssl/apiserver-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountCrt,
			Path:         "/etc/kubernetes/ssl/service-account-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountCA,
			Path:         "/etc/kubernetes/ssl/service-account-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountKey,
			Path:         "/etc/kubernetes/ssl/service-account-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/server-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/server-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/server-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		// Add second copy of files for etcd client certs. Will be replaced by
		// a separate client cert.
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     k8scloudconfig.GzipBase64,
			Permissions:  FilePermission,
		},
		{
			AssetContent: waitDockerConfTemplate,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.keys.APIServerEncryptionKey,
			Path:         "/etc/kubernetes/config/k8s-encryption-config.yaml",
			Owner:        FileOwner,
			Permissions:  0644,
		},
	}

	var newFiles []k8scloudconfig.FileAsset

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, e.customObject.Spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := k8scloudconfig.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *v_0_1_0MasterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: decryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: masterFormatVarLibDockerServiceTemplate,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: ephemeralVarLibDockerMountTemplate,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, e.customObject.Spec)
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

func (e *v_0_1_0MasterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	newSections := []k8scloudconfig.VerbatimSection{
		{
			Name:    "storage",
			Content: instanceStorageTemplate,
		},
		{
			Name:    "storageclass",
			Content: instanceStorageClassTemplate,
		},
	}
	return newSections
}
