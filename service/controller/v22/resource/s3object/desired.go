package s3object

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	{
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
			Bucket: key.BucketName(customObject, cc.Status.Cluster.AWSAccount.ID),
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
			Bucket: key.BucketName(customObject, cc.Status.Cluster.AWSAccount.ID),
			Body:   b,
			Key:    k,
		}
	}

	return output, nil
}
