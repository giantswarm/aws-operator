package create

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/k8scloudconfig"
)

var (
	filesMeta []cloudconfig.FileMetadata = []cloudconfig.FileMetadata{
		cloudconfig.FileMetadata{
			AssetContent: decryptTLSAssetsScriptTemplate,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner:        "root:root",
			Permissions:  0700,
		},
		cloudconfig.FileMetadata{
			AssetContent: calicoEnvironmentFileTemplate,
			Path:         "/etc/calico-environment",
			Owner:        "root:root",
			Permissions:  0644,
		},
	}
	unitsMeta []cloudconfig.UnitMetadata = []cloudconfig.UnitMetadata{
		cloudconfig.UnitMetadata{
			AssetContent: decryptTLSAssetsServiceTemplate,
			Name:         "decrypt-tls-assets.service",
			Enable:       true,
			Command:      "start",
		},
	}
)

var (
	assetTemplates = map[string]string{
		prefixMaster: cloudconfig.MasterTemplate,
		prefixWorker: cloudconfig.WorkerTemplate,
	}
)

type CloudConfigExtension struct {
	AwsInfo awstpr.Spec
}

func NewCloudConfigExtension(awsSpec awstpr.Spec) *CloudConfigExtension {
	return &CloudConfigExtension{
		AwsInfo: awsSpec,
	}
}

func (c *CloudConfigExtension) Files() ([]cloudconfig.FileAsset, error) {
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

func (c *CloudConfigExtension) Units() ([]cloudconfig.UnitAsset, error) {
	units := make([]cloudconfig.UnitAsset, 0, len(unitsMeta))

	for _, unitMeta := range unitsMeta {
		content, err := cloudconfig.RenderAssetContent(unitMeta.AssetContent, c.AwsInfo)
		if err != nil {
			return nil, err
		}

		unitAsset := cloudconfig.UnitAsset{
			Metadata: unitMeta,
			Content:  content,
		}

		units = append(units, unitAsset)
	}

	return units, nil
}

func (s *Service) cloudConfig(prefix string, params cloudconfig.CloudConfigTemplateParams, awsSpec awstpr.Spec) (string, error) {
	extension := NewCloudConfigExtension(awsSpec)

	cloudconfig, err := cloudconfig.NewCloudConfig(assetTemplates[prefix], params, extension)
	if err != nil {
		return "", err
	}

	if err := cloudconfig.ExecuteTemplate(); err != nil {
		return "", err
	}

	return cloudconfig.Base64(), nil
}
