package create

import (
	"bytes"
	"encoding/base64"
	"html/template"

	microerror "github.com/giantswarm/microkit/error"
)

type SmallCloudconfigConfig struct {
	MachineType string
	Region      string
	S3DirURI    string
}

func (s *Service) SmallCloudconfig(config SmallCloudconfigConfig) (string, error) {
	tmpl, err := template.New("smallCloudconfig").Parse(userDataScriptTemplate)
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, config)
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
