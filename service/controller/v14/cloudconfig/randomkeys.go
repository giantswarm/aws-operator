package cloudconfig

import (
	"bytes"
	"context"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v13/templates/cloudconfig"
)

func renderRandomKeyTmplSet(ctx context.Context, encrypter encrypter.Interface, key string, clusterKeys randomkeys.Cluster) (RandomKeyTmplSet, error) {
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

		enc, err := encrypter.Encrypt(ctx, key, buf.String())
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}

		com, err := compactor([]byte(enc))
		if err != nil {
			return RandomKeyTmplSet{}, microerror.Mask(err)
		}

		randomKeyTmplSet.APIServerEncryptionKey = com
	}

	return randomKeyTmplSet, nil
}
