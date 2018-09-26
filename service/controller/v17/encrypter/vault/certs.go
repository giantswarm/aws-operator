package vault

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v17/encrypter"
)

func (e *Encrypter) EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error) {
	rawTLS := encrypter.CreateRawTLSAssets(assets)

	key, err := e.EncryptionKey(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	encTLS, err := rawTLS.Encrypt(ctx, e, key)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	compTLS, err := encTLS.Compact()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return compTLS, nil
}
