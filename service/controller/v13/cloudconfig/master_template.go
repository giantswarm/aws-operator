package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_3_4"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v13/templates/cloudconfig"
)

// NewMasterTemplate generates a new master cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(customObject v1alpha1.AWSConfig, certs legacy.CompactTLSAssets, clusterKeys randomkeys.Cluster, kmsKeyARN string) (string, error) {
	var err error

	randomKeyTmplSet, err := renderRandomKeyTmplSet(c.kmsClient, clusterKeys, kmsKeyARN)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.DisableEncryptionAtREST = true
		params.DisableIngressController = true
		params.EtcdPort = customObject.Spec.Cluster.Etcd.Port
		params.Extension = &MasterExtension{
			certs:            certs,
			customObject:     customObject,
			RandomKeyTmplSet: randomKeyTmplSet,
		}
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = c.k8sAPIExtraArgs
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = c.k8sKubeletExtraArgs
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

// RandomKeyTmplSet holds a collection of rendered templates for random key
// encryption via KMS.
type RandomKeyTmplSet struct {
	APIServerEncryptionKey string
}

type MasterExtension struct {
	certs            legacy.CompactTLSAssets
	customObject     v1alpha1.AWSConfig
	RandomKeyTmplSet RandomKeyTmplSet
}

func (e *MasterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
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
			AssetContent: e.RandomKeyTmplSet.APIServerEncryptionKey,
			Path:         "/etc/kubernetes/encryption/k8s-encryption-config.yaml.enc",
			Owner:        FileOwner,
			Encoding:     GzipBase64Encoding,
			Permissions:  0644,
		},
		{
			AssetContent: cloudconfig.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsScript,
			Path:         "/opt/bin/decrypt-keys-assets",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Add use-proxy-protocol to ingress-controller ConfigMap, this doesn't work
		// on KVM because of dependencies on hardware LB configuration.
		{
			AssetContent: cloudconfig.IngressControllerConfigMap,
			Path:         "/srv/ingress-controller-cm.yml",
			Owner:        FileOwner,
			Permissions:  0644,
		},
		// NVME disks udev rules and script.
		// Workaround for https://github.com/coreos/bugs/issues/2399
		{
			AssetContent: cloudconfig.NVMEUdevRule,
			Path:         "/etc/udev/rules.d/10-ebs-nvme-mapping.rules",
			Owner:        FileOwner,
			Permissions:  0644,
		},
		{
			AssetContent: cloudconfig.NVMEUdevScript,
			Path:         "/opt/ebs-nvme-mapping",
			Owner:        FileOwner,
			Permissions:  0766,
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
		// Create symlinks for nvme disks.
		// This service should be started only on first boot.
		{
			AssetContent: cloudconfig.NVMEUdevTriggerUnit,
			Name:         "ebs-nvme-udev-trigger.service",
			Enable:       false,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			// Do not enable TLS assets decrypt unit so that it won't get automatically
			// executed on master reboot. This will prevent eventual races with the
			// asset files creation.
			Enable:  false,
			Command: "start",
		},
		{
			AssetContent: cloudconfig.MasterFormatVarLibDockerService,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsService,
			Name:         "decrypt-keys-assets.service",
			// Do not enable key decrypt unit so that it won't get automatically
			// executed on master reboot. This will prevent eventual races with the
			// key files creation.
			Enable:  false,
			Command: "start",
		},
		// Format etcd EBS volume.
		{
			AssetContent: cloudconfig.FormatEtcdVolume,
			Name:         "format-etcd-ebs.service",
			Enable:       true,
			Command:      "start",
		},
		// Mount etcd EBS volume.
		{
			AssetContent: cloudconfig.MountEtcdVolume,
			Name:         "var-lib-etcd.mount",
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
			Content: cloudconfig.InstanceStorage,
		},
		{
			Name:    "storageclass",
			Content: cloudconfig.InstanceStorageClass,
		},
	}
	return newSections
}
