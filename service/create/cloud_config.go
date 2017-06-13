package create

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
)

var (
	assetTemplates = map[string]string{
		prefixMaster: cloudconfig.MasterTemplate,
		prefixWorker: cloudconfig.WorkerTemplate,
	}
)

type CloudConfigExtension struct {
	AwsInfo   awstpr.Spec
	TLSAssets *certificatetpr.CompactTLSAssets
}

func (c *CloudConfigExtension) renderFiles(filesMeta []cloudconfig.FileMetadata) ([]cloudconfig.FileAsset, error) {
	files := make([]cloudconfig.FileAsset, 0, len(filesMeta))

	for _, fileMeta := range filesMeta {
		content, err := cloudconfig.RenderAssetContent(fileMeta.AssetContent, c.AwsInfo)
		if err != nil {
			return nil, err
		}

		fileAsset := cloudconfig.FileAsset{
			Metadata: fileMeta,
			Content:  content,
		}

		files = append(files, fileAsset)
	}

	return files, nil
}

func (c *CloudConfigExtension) renderUnits(unitsMeta []cloudconfig.UnitMetadata) ([]cloudconfig.UnitAsset, error) {
	units := make([]cloudconfig.UnitAsset, 0, len(unitsMeta))

	for _, unitMeta := range unitsMeta {
		content, err := cloudconfig.RenderAssetContent(unitMeta.AssetContent, c.AwsInfo)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		unitAsset := cloudconfig.UnitAsset{
			Metadata: unitMeta,
			Content:  content,
		}

		units = append(units, unitAsset)
	}

	return units, nil
}

type MasterCloudConfigExtension struct {
	CloudConfigExtension
}

func NewMasterCloudConfigExtension(awsSpec awstpr.Spec, tlsAssets *certificatetpr.CompactTLSAssets) *MasterCloudConfigExtension {
	return &MasterCloudConfigExtension{
		CloudConfigExtension{
			AwsInfo:   awsSpec,
			TLSAssets: tlsAssets,
		},
	}
}

func (m *MasterCloudConfigExtension) Files() ([]cloudconfig.FileAsset, error) {
	masterFilesMeta := []cloudconfig.FileMetadata{
		cloudconfig.FileMetadata{
			AssetContent: decryptTLSAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        "root:root",
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.APIServerCrt,
			Path:         "/etc/kubernetes/ssl/apiserver-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.APIServerCA,
			Path:         "/etc/kubernetes/ssl/apiserver-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.APIServerKey,
			Path:         "/etc/kubernetes/ssl/apiserver-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.ServiceAccountCrt,
			Path:         "/etc/kubernetes/ssl/service-account-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.ServiceAccountCA,
			Path:         "/etc/kubernetes/ssl/service-account-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.ServiceAccountKey,
			Path:         "/etc/kubernetes/ssl/service-account-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/server-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/server-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/server-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		// Add second copy of files for etcd client certs. Will be replaced by
		// a separate client cert.
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: m.TLSAssets.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: waitDockerConfTemplate,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        "root:root",
			Permissions:  0700,
		},
	}

	files, err := m.renderFiles(masterFilesMeta)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return files, nil
}

func (m *MasterCloudConfigExtension) Units() ([]cloudconfig.UnitAsset, error) {
	unitsMeta := []cloudconfig.UnitMetadata{
		cloudconfig.UnitMetadata{
			AssetContent: decryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
			Command:      "start",
		},
		cloudconfig.UnitMetadata{
			AssetContent: varLibDockerMountTemplate,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
	}

	units, err := m.renderUnits(unitsMeta)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return units, nil
}

type WorkerCloudConfigExtension struct {
	CloudConfigExtension
}

func NewWorkerCloudConfigExtension(awsSpec awstpr.Spec, tlsAssets *certificatetpr.CompactTLSAssets) *WorkerCloudConfigExtension {
	return &WorkerCloudConfigExtension{
		CloudConfigExtension{
			AwsInfo:   awsSpec,
			TLSAssets: tlsAssets,
		},
	}
}

func (w *WorkerCloudConfigExtension) Files() ([]cloudconfig.FileAsset, error) {
	workerFilesMeta := []cloudconfig.FileMetadata{
		cloudconfig.FileMetadata{
			AssetContent: decryptTLSAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        "root:root",
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.WorkerCrt,
			Path:         "/etc/kubernetes/ssl/worker-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.WorkerCA,
			Path:         "/etc/kubernetes/ssl/worker-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.WorkerKey,
			Path:         "/etc/kubernetes/ssl/worker-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.CalicoClientCrt,
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.CalicoClientCA,
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.CalicoClientKey,
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.EtcdServerCrt,
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.EtcdServerCA,
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: w.TLSAssets.EtcdServerKey,
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			Owner:        "root:root",
			Encoding:     cloudconfig.GzipBase64,
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: waitDockerConfTemplate,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner:        "root:root",
			Permissions:  0700,
		},
	}

	files, err := w.renderFiles(workerFilesMeta)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return files, nil
}

func (w *WorkerCloudConfigExtension) Units() ([]cloudconfig.UnitAsset, error) {
	unitsMeta := []cloudconfig.UnitMetadata{
		cloudconfig.UnitMetadata{
			AssetContent: decryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
			Command:      "start",
		},
		cloudconfig.UnitMetadata{
			AssetContent: formatVarLibDockerServiceTemplate,
			Name:         "format-var-lib-docker.service",
			Enable:       true,
			Command:      "start",
		},
		cloudconfig.UnitMetadata{
			AssetContent: varLibDockerMountTemplate,
			Name:         "var-lib-docker.mount",
			Enable:       true,
			Command:      "start",
		},
	}

	units, err := w.renderUnits(unitsMeta)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return units, nil
}

// VerbatimSections is defined on CloudConfigExtension since there's no difference
// between master and workers sections.
func (c *CloudConfigExtension) VerbatimSections() []cloudconfig.VerbatimSection {
	sections := []cloudconfig.VerbatimSection{
		{
			Name:    "storage",
			Content: instanceStorageTemplate,
		},
	}

	return sections
}

func (s *Service) cloudConfig(prefix string, params cloudconfig.Params, awsSpec awstpr.Spec, tlsAssets *certificatetpr.CompactTLSAssets) (string, error) {
	var template string
	switch prefix {
	case prefixMaster:
		template = cloudconfig.MasterTemplate
	case prefixWorker:
		template = cloudconfig.WorkerTemplate
	default:
		return "", invalidCloudconfigExtensionNameError
	}

	cc, err := cloudconfig.NewCloudConfig(template, params)
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	if err := cc.ExecuteTemplate(); err != nil {
		return "", microerror.MaskAny(err)
	}

	return cc.Base64(), nil
}
