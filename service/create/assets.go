package create

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certificatetpr"
	microerror "github.com/giantswarm/microkit/error"
)

func (s *Service) encodeTLSAssets(assets certificatetpr.AssetsBundle, svc *kms.KMS, kmsKeyArn string) (*certificatetpr.CompactTLSAssets, error) {
	rawTLS := createRawTLSAssets(assets)

	encTLS, err := rawTLS.encrypt(svc, kmsKeyArn)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	compTLS, err := encTLS.compact()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return compTLS, nil
}
