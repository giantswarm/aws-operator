package s3objectv3

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

func (s *Resource) encodeTLSAssets(assets legacy.AssetsBundle, kmsKeyArn string) (*legacy.CompactTLSAssets, error) {
	rawTLS := createRawTLSAssets(assets)

	encTLS, err := rawTLS.encrypt(s.awsClients.KMS, kmsKeyArn)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	compTLS, err := encTLS.compact()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return compTLS, nil
}

func (s *Resource) encodeKeyAssets(assets map[randomkeytpr.Key][]byte, kmsKeyArn string) (*randomkeytpr.CompactRandomKeyAssets, error) {
	encryptionKey, ok := assets[randomkeytpr.EncryptionKey]
	if !ok {
		return nil, microerror.Mask(invalidConfigError)
	}

	encryptionConfig, err := s.EncryptionConfig(string(encryptionKey))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	rawKeys := make(rawKeyAssets)
	rawKeys[randomkeytpr.EncryptionKey] = []byte(encryptionConfig)

	encKeys, err := rawKeys.encrypt(s.awsClients.KMS, kmsKeyArn)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	compKeys, err := encKeys.compact()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return compKeys, nil
}

func (s *Resource) EncryptionConfig(encryptionKey string) (string, error) {
	tmpl, err := template.New("encryptionConfig").Parse(encryptionConfigTemplate)
	if err != nil {
		return "", microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(buf.Bytes()), nil
}
