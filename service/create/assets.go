package create

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
)

func (s *Service) encodeTLSAssets(awsSession *session.Session, kmsKeyArn string) (*cloudconfig.CompactTLSAssets, error) {
	rawTLS, err := readRawTLSAssets(s.certsDir)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	encTLS, err := rawTLS.encrypt(awsSession, kmsKeyArn)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	compTLS, err := encTLS.compact()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	return compTLS, nil
}
