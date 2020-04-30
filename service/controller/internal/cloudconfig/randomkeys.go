package cloudconfig

import (
	"bytes"
	"context"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/internal/templates/cloudconfig"
)

func renderRandomKeyTmplSet(ctx context.Context, encrypter encrypter.Interface, key string, clusterKeys randomkeys.Cluster) (RandomKeyTmplSet, string, error) {
	var unencrypted string
	var randomKeyTmplSet RandomKeyTmplSet
	{
		tmpl, err := template.New("encryption-config").Parse(cloudconfig.EncryptionConfig)
		if err != nil {
			return RandomKeyTmplSet{}, "", microerror.Mask(err)
		}
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, struct {
			EncryptionKey string
		}{
			EncryptionKey: string(clusterKeys.APIServerEncryptionKey),
		})
		if err != nil {
			return RandomKeyTmplSet{}, "", microerror.Mask(err)
		}

		unencrypted = buf.String()
		enc, err := encrypter.Encrypt(ctx, key, unencrypted)
		if err != nil {
			return RandomKeyTmplSet{}, "", microerror.Mask(err)
		}

		randomKeyTmplSet.APIServerEncryptionKey = enc
	}

	return randomKeyTmplSet, unencrypted, nil
}
