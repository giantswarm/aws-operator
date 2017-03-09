package create

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"io/ioutil"

	microerror "github.com/giantswarm/microkit/error"
)

var (
	smallCloudconfigPath = "templates/user-data.sh"
)

type smallCloudconfigProvider interface {
	smallCloudconfigContent() ([]byte, error)
}

type fsSmallCloudconfigProvider struct {
	smallCloudconfigFile string
}

func newFsSmallCloudconfigProvider(file string) *fsSmallCloudconfigProvider {
	return &fsSmallCloudconfigProvider{
		smallCloudconfigFile: file,
	}
}

func (f *fsSmallCloudconfigProvider) smallCloudconfigContent() ([]byte, error) {
	return ioutil.ReadFile(f.smallCloudconfigFile)
}

type SmallCloudconfigConfig struct {
	MachineType string
	Region      string
	S3Bucket    string
	ClusterID   string
}

func (s *Service) SmallCloudconfig(provider smallCloudconfigProvider, config SmallCloudconfigConfig) (string, error) {
	b, err := provider.smallCloudconfigContent()
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	tmpl, err := template.New("smallCloudconfig").Parse(string(b))
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
