package create

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"text/template"

	microerror "github.com/giantswarm/microkit/error"
)

type CloudConfig struct {
	config   string
	params   cloudconfigTemplateParams
	path     string
	template string
}

func newCloudConfig(templatePath string, params cloudconfigTemplateParams) (*CloudConfig, error) {
	cloudConfigTempl, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return &CloudConfig{
		path:     templatePath,
		template: string(cloudConfigTempl),
		params:   params,
	}, nil
}

func (c *CloudConfig) executeTemplate() error {
	tmpl, err := template.New(c.path).Parse(c.template)
	if err != nil {
		return microerror.MaskAny(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, c.params)
	if err != nil {
		return microerror.MaskAny(err)
	}
	c.config = buf.String()

	return nil
}

func (c *CloudConfig) base64() string {
	cloudConfigBytes := []byte(c.config)

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(cloudConfigBytes)
	w.Close()

	return base64.StdEncoding.EncodeToString(b.Bytes())
}
