package kms

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v15/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v15/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

func (k *Encrypter) EncryptTLSAssets(ctx context.Context, customObject v1alpha1.AWSConfig, assets legacy.AssetsBundle) (*legacy.CompactTLSAssets, error) {
	rawTLS := encrypter.CreateRawTLSAssets(assets)

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	kmsKeyARN, err := sc.AWSService.GetKeyArn(clusterID)

	encTLS, err := rawTLS.Encrypt(ctx, k, kmsKeyARN)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	compTLS, err := encTLS.Compact()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return compTLS, nil
}
