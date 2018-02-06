package adapter

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"os"
	"path/filepath"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v3/key"
)

func SmallCloudconfig(config SmallCloudconfigConfig) (string, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return "", microerror.Mask(err)
	}
	rootDir, err := key.RootDir(baseDir, RootDirElement)
	if err != nil {
		return "", microerror.Mask(err)
	}
	templateFile := filepath.Join(rootDir, smallCloudConfigTemplate)

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
