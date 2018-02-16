package v_3_1_1

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"

	"github.com/giantswarm/microerror"
)

type CloudConfigConfig struct {
	Params   Params
	Template string
}

func DefaultCloudConfigConfig() CloudConfigConfig {
	return CloudConfigConfig{
		Params:   Params{},
		Template: "",
	}
}

type CloudConfig struct {
	config   string
	params   Params
	template string
}

func NewCloudConfig(config CloudConfigConfig) (*CloudConfig, error) {
	if err := config.Params.Validate(); err != nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Params.%s", err)
	}
	if config.Template == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Template must not be empty")
	}

	// Set default params.
	if config.Params.MasterAPIDomain == "" {
		config.Params.MasterAPIDomain = config.Params.Cluster.Kubernetes.API.Domain
	}
	if config.Params.Hyperkube.Apiserver.BindAddress == "" {
		config.Params.Hyperkube.Apiserver.BindAddress = defaultHyperkubeApiserverBindAddress
	}
	// Default to 443 for non AWS providers.
	if config.Params.EtcdPort == 0 {
		config.Params.EtcdPort = 443
	}

	c := &CloudConfig{
		config:   "",
		params:   config.Params,
		template: config.Template,
	}

	return c, nil
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
