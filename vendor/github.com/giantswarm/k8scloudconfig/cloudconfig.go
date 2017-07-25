package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
)

type Params struct {
	Cluster   clustertpr.Spec
	Extension Extension
	Node      spec.Node
}

type CloudConfig struct {
	config   string
	params   Params
	template string
}

func NewCloudConfig(template string, params Params) (*CloudConfig, error) {
	newCloudConfig := &CloudConfig{
		config:   "",
		params:   params,
		template: template,
	}

	return newCloudConfig, nil
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

func (c *CloudConfig) String() string {
	return c.config
}
