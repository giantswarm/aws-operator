package create

import (
	"bytes"
	"encoding/base64"
	"html/template"

	"github.com/giantswarm/microerror"
)

type SmallCloudconfigConfig struct {
	MachineType string
	Region      string
	S3URI       string
}

func (s *Service) SmallCloudconfig(config SmallCloudconfigConfig) (string, error) {
	tmpl, err := template.New("smallCloudconfig").Parse(userDataScriptTemplate)
	if err != nil {
		return "", microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
