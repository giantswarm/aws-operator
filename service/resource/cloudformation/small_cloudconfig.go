package cloudformation

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"path/filepath"

	"github.com/giantswarm/microerror"
)

func SmallCloudconfig(config SmallCloudconfigConfig) (string, error) {
	templateFile, err := filepath.Abs(filepath.Join("../../../", smallCloudConfigTemplate))
	if err != nil {
		return "", microerror.Mask(err)
	}

	tmpl, err := template.ParseFiles(templateFile)
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
