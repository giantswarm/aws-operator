package legacy

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

func (s *Resource) encodeTLSAssets(assets certificatetpr.AssetsBundle, svc *kms.KMS, kmsKeyArn string) (*certificatetpr.CompactTLSAssets, error) {
	rawTLS := createRawTLSAssets(assets)

	encTLS, err := rawTLS.encrypt(svc, kmsKeyArn)
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
		return nil, microerror.Maskf(executionFailedError, "could not get encryption keys from secrets")
	}

	encryptionConfig, err := s.EncryptionConfig(string(encryptionKey))

	if err != nil {
		return nil, microerror.Maskf(executionFailedError, fmt.Sprintf("could not generate encryption config: '%#v'", err))
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
