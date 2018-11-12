package s3object

import (
	"context"
	"fmt"

	"github.com/giantswarm/aws-operator/service/controller/v18patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v18patch1/key"
	"github.com/giantswarm/certs"
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

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("labels: %#v", customObject.Spec.Cluster.Kubernetes.Kubelet.Labels))

	{
		b, err := r.cloudConfig.NewMasterTemplate(ctx, customObject, clusterCerts, clusterKeys)
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
		b, err := r.cloudConfig.NewWorkerTemplate(ctx, customObject, clusterCerts)
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
