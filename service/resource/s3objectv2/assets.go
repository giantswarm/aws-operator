package s3objectv2

import (
	"bytes"
	"html/template"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

func (s *Resource) encodeTLSAssets(assets certificatetpr.AssetsBundle, kmsKeyArn string) (*certificatetpr.CompactTLSAssets, error) {
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

func (s *Resource) encodeKeyAssets(assets map[randomkeytpr.Key][]byte, svc *kms.KMS, kmsKeyArn string) (*randomkeytpr.CompactRandomKeyAssets, error) {

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

	encKeys, err := rawKeys.encrypt(svc, kmsKeyArn)
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
