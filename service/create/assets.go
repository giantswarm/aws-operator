package create

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
)

func (s *Service) encodeTLSAssets(assets certificatetpr.AssetsBundle, svc *kms.KMS, kmsKeyArn string) (*certificatetpr.CompactTLSAssets, error) {
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
