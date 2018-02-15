package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_1_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"

	"github.com/giantswarm/aws-operator/service/awsconfig/v5/templates/cloudconfig"
)

const (
	// MasterCloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	MasterCloudConfigVersion = "v_3_1_0"
)

// NewMasterTemplate generates a new master cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(customObject v1alpha1.AWSConfig, certs legacy.CompactTLSAssets, keys randomkeytpr.CompactRandomKeyAssets) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.ApiserverEncryptionKey = keys.APIServerEncryptionKey
		params.Cluster = customObject.Spec.Cluster
		params.EtcdPort = customObject.Spec.Cluster.Etcd.Port
		params.Extension = &MasterExtension{
			certs:        certs,
			customObject: customObject,
			keys:         keys,
		}
		params.Hyperkube.Apiserver.Docker.CommandExtraArgs = c.k8sAPIExtraArgs
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = k8scloudconfig.MasterTemplate

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

type MasterExtension struct {
	certs        legacy.CompactTLSAssets
	customObject v1alpha1.AWSConfig
	keys         randomkeytpr.CompactRandomKeyAssets
}

func (e *MasterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerCrt,
			Path:         "/etc/kubernetes/ssl/apiserver-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerCA,
			Path:         "/etc/kubernetes/ssl/apiserver-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.APIServerKey,
			Path:         "/etc/kubernetes/ssl/apiserver-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountCrt,
			Path:         "/etc/kubernetes/ssl/service-account-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountCA,
			Path:         "/etc/kubernetes/ssl/service-account-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.ServiceAccountKey,
			Path:         "/etc/kubernetes/ssl/service-account-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/server-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/server-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/server-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		// Add second copy of files for etcd client certs. Will be replaced by
		// a separate client cert.
		{
			AssetContent: e.certs.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.certs.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  FilePermission,
		},
		{
			AssetContent: waitDockerConfTemplate,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-keys-assets",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: e.keys.APIServerEncryptionKey,
			Path:         "/etc/kubernetes/encryption/k8s-encryption-config.yaml.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  0644,
		},
		// Add use-proxy-protocol to ingress-controller ConfigMap, this doesn't work
		// on KVM because of dependencies on hardware LB configuration.
		{
			AssetContent: ingressControllerConfigMapTemplate,
			Path:         "/srv/ingress-controller-cm.yml",
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

func (e *MasterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			// Do not enable TLS assets decrypt unit so that it won't get automatically
			// executed on master reboot. This will prevent eventual races with the
			// asset files creation.
			Enable:  false,
			Command: "start",
		},
		{
			AssetContent: cloudconfig.MasterFormatVarLibDockerServiceTemplate,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibDockerMountTemplate,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsServiceTemplate,
			Name:         "decrypt-keys-assets.service",
			// Do not enable key decrypt unit so that it won't get automatically
			// executed on master reboot. This will prevent eventual races with the
			// key files creation.
			Enable:  false,
			Command: "start",
		},
		// Format etcd EBS volume.
		{
			AssetContent: formatEtcdVolume,
			Name:         "format-etcd-ebs.service",
			Enable:       true,
			Command:      "start",
		},
		// Mount etcd EBS volume.
		{
			AssetContent: mountEtcdVolume,
			Name:         "etc-kubernetes-data-etcd.mount",
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

func (e *MasterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
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
