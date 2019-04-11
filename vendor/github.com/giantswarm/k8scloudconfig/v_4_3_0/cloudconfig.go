package v_4_3_0

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"

	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"
)

const (
	defaultRegistryDomain = "quay.io"
	kubernetesImage       = "giantswarm/hyperkube:v1.13.4"
	etcdImage             = "giantswarm/etcd:v3.3.12"
	etcdPort              = 443
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

func DefaultParams() Params {
	return Params{
		EtcdPort:       etcdPort,
		RegistryDomain: defaultRegistryDomain,
		Images: Images{
			Kubernetes: kubernetesImage,
			Etcd:       etcdImage,
		},
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
		return microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, c.params)
	if err != nil {
		return microerror.Mask(err)
	}

	ignitionJSON, err := ignition.ConvertTemplatetoJSON(buf.Bytes())
	if err != nil {
		return microerror.Mask(err)
	}

	c.config = string(ignitionJSON)

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
