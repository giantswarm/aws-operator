package cloudconfig

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v12patch1/templates/cloudconfig"
)

func renderRandomKeyTmplSet(kmsClient KMSClient, clusterKeys randomkeys.Cluster, kmsKeyARN string) (RandomKeyTmplSet, error) {
	var randomKeyTmplSet RandomKeyTmplSet
	{
		tmpl, err := template.New("encryption-config").Parse(cloudconfig.EncryptionConfig)
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, struct {
			EncryptionKey string
		}{
			EncryptionKey: string(clusterKeys.APIServerEncryptionKey),
		})
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}

		enc, err := encryptor(kmsClient, kmsKeyARN, buf.Bytes())
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}

		com, err := compactor(enc)
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}

		randomKeyTmplSet.APIServerEncryptionKey = com
	}

	return randomKeyTmplSet, nil
}
