package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/node"
)

type CloudConfigTemplateParams struct {
	Cluster   clustertpr.Cluster
	Node      node.Node
	TLSAssets CompactTLSAssets
	Files     []FileAsset
	Units     []UnitAsset
}

type CloudConfig struct {
	config    string
	extension OperatorExtension
	params    CloudConfigTemplateParams
	template  string
}

func NewCloudConfig(template []byte, params CloudConfigTemplateParams, extension OperatorExtension) (*CloudConfig, error) {
	files, err := extension.Files()
	if err != nil {
		return nil, err
	}

	units, err := extension.Units()
	if err != nil {
		return nil, err
	}

	params.Files = files
	params.Units = units

	return &CloudConfig{
		template: string(template),
		params:   params,
	}, nil
}

func (c *CloudConfig) ExecuteTemplate() error {
	tmpl, err := template.New("cloudconfig").Parse(c.template)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, c.params)
	if err != nil {
		return err
	}
	c.config = buf.String()

	return nil
}

func (c *CloudConfig) Base64() string {
	cloudConfigBytes := []byte(c.config)

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(cloudConfigBytes)
	w.Close()

	return base64.StdEncoding.EncodeToString(b.Bytes())
}
