package s3object

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v21/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v21/key"
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
	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	{
		accountID, err = sc.AWSService.GetAccountID()
		if err != nil {
			return nil, microerror.Mask(err)
		}
		clusterCerts, err = r.certsSearcher.SearchCluster(key.ClusterID(customObject))
		if err != nil {
			return nil, microerror.Mask(err)
		}
		clusterKeys, err = r.randomKeysSearcher.SearchCluster(key.ClusterID(customObject))
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	output := map[string]BucketObjectState{}

	{
		b, err := r.cloudConfig.NewMasterTemplate(ctx, customObject, clusterCerts, clusterKeys)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		k := key.BucketObjectName(customObject, key.KindMaster)
		output[k] = BucketObjectState{
			Bucket: key.BucketName(customObject, accountID),
			Body:   b,
			Key:    k,
		}
	}

	{
		b, err := r.cloudConfig.NewWorkerTemplate(ctx, customObject, clusterCerts)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		k := key.BucketObjectName(customObject, key.KindWorker)
		output[k] = BucketObjectState{
			Bucket: key.BucketName(customObject, accountID),
			Body:   b,
			Key:    k,
		}
	}

	return output, nil
}
