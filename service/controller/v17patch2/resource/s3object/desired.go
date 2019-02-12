package s3object

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/v17patch2/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v17patch2/key"
	"github.com/giantswarm/legacycerts/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = r.encrypter.EncryptionKey(ctx, customObject)
	if r.encrypter.IsKeyNotFound(err) && key.IsDeleted(customObject) {
		// we can get here during deletion, if the key is already deleted we can safely exit.
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var accountID string
	var certs legacy.AssetsBundle
	var tlsAssets *legacy.CompactTLSAssets
	var clusterKeys randomkeys.Cluster
	{
		accountID, err = sc.AWSService.GetAccountID()
		if err != nil {
			return nil, microerror.Mask(err)
		}
		certs, err = r.certWatcher.SearchCerts(key.ClusterID(customObject))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		tlsAssets, err = r.encrypter.EncryptTLSAssets(ctx, customObject, certs)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		clusterKeys, err = r.randomKeySearcher.SearchCluster(key.ClusterID(customObject))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	output := map[string]BucketObjectState{}

	{
		b, err := r.cloudConfig.NewMasterTemplate(ctx, customObject, *tlsAssets, clusterKeys)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		k := key.BucketObjectName(customObject, prefixMaster)
		output[k] = BucketObjectState{
			Bucket: key.BucketName(customObject, accountID),
			Body:   b,
			Key:    k,
		}
	}

	{
		b, err := r.cloudConfig.NewWorkerTemplate(ctx, customObject, *tlsAssets)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		k := key.BucketObjectName(customObject, prefixWorker)
		output[k] = BucketObjectState{
			Bucket: key.BucketName(customObject, accountID),
			Body:   b,
			Key:    k,
		}
	}

	return output, nil
}
