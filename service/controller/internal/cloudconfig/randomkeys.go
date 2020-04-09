package cloudconfig

import (
	"bytes"
	"context"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	cloudconfigtemplate "github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig/template"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
)

// randomKeyTmplSet holds a collection of rendered templates for random key
// encryption via KMS.
type RandomKeyTmplSet struct {
	APIServerEncryptionKey string
}

func renderRandomKeyTmplSet(ctx context.Context, encrypter encrypter.Interface, key string, clusterKeys randomkeys.Cluster) (RandomKeyTmplSet, error) {
	var randomKeyTmplSet RandomKeyTmplSet
	{
		tmpl, err := template.New("encryption-config").Parse(cloudconfigtemplate.EncryptionConfig)
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

		randomKeyTmplSet.APIServerEncryptionKey = enc
	}

	return randomKeyTmplSet, nil
}
