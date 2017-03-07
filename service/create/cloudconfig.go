package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/k8scloudconfig"

	"github.com/giantswarm/aws-operator/bindata"
)

var (
	filesMeta []cloudconfig.FileMetadata = []cloudconfig.FileMetadata{
		cloudconfig.FileMetadata{
			AssetPath:   "templates/decrypt-tls-assets",
			Path:        "/opt/bin/decrypt-tls-assets",
			Owner:       "root:root",
			Permissions: 0700,
		},
	}
	unitsMeta []cloudconfig.UnitMetadata = []cloudconfig.UnitMetadata{
		cloudconfig.UnitMetadata{
			AssetPath: "templates/decrypt-tls-assets.service",
			Name:      "decrypt-tls-assets.service",
			Enable:    true,
			Command:   "start",
		},
	}
)

type CloudConfigExtension struct {
	AwsInfo awstpr.Spec
}

func NewCloudConfigExtension() *CloudConfigExtension {
	return &CloudConfigExtension{}
}

func (c *CloudConfigExtension) Files() ([]cloudconfig.FileAsset, error) {
	files := make([]cloudconfig.FileAsset, 0, len(filesMeta))

	for _, fileMeta := range filesMeta {
		rawContent, err := bindata.Asset(fileMeta.AssetPath)
		if err != nil {
			return nil, err
		}

		content, err := cloudconfig.RenderAssetContent(rawContent, c.AwsInfo)
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
		rawContent, err := bindata.Asset(unitMeta.AssetPath)
		if err != nil {
			return nil, err
		}

		content, err := cloudconfig.RenderAssetContent(rawContent, c.AwsInfo)
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

func (s *Service) cloudConfig(prefix string, params cloudconfig.CloudConfigTemplateParams) (string, error) {
	template, err := cloudconfig.Asset(fmt.Sprintf("templates/%s.yaml", prefix))
	if err != nil {
		return "", err
	}

	extension := NewCloudConfigExtension()

	cloudconfig, err := cloudconfig.NewCloudConfig(template, params, extension)
	if err != nil {
		return "", err
	}

	if err := cloudconfig.ExecuteTemplate(); err != nil {
		return "", err
	}

	return cloudconfig.Base64(), nil
}
