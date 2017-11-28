package adapter

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"path/filepath"
	"runtime"

	"github.com/giantswarm/microerror"
)

func SmallCloudconfig(config SmallCloudconfigConfig) (string, error) {
	_, filename, _, _ := runtime.Caller(1)
	templateFile, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "../../../../", smallCloudConfigTemplate))
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
