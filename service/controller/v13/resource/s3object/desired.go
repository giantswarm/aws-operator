package s3object

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v13/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v13/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	output := map[string]BucketObjectState{}
	if err != nil {
		return output, microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := sc.AWSService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)

	if IsKeyNotFound(err) {
		// we can get here during deletion, if the key is already deleted we can safely exit.
		return output, nil
	}
	if err != nil {
		return output, microerror.Mask(err)
	}

	certs, err := r.certWatcher.SearchCerts(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	tlsAssets, err := r.encrypter.EncryptTLSAssets(ctx, customObject, certs)
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterKeys, err := r.randomKeySearcher.SearchCluster(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	masterBody, err := sc.CloudConfig.NewMasterTemplate(ctx, customObject, *tlsAssets, clusterKeys)
	if err != nil {
		return output, microerror.Mask(err)
	}

	masterObjectName := key.BucketObjectName(cloudconfig.CloudConfigVersion, prefixMaster)
	masterCloudConfig := BucketObjectState{
		Bucket: key.BucketName(customObject, accountID),
		Body:   masterBody,
		Key:    masterObjectName,
	}
	output[masterObjectName] = masterCloudConfig

	workerBody, err := sc.CloudConfig.NewWorkerTemplate(ctx, customObject, *tlsAssets)
	if err != nil {
		return output, microerror.Mask(err)
	}

	workerObjectName := key.BucketObjectName(cloudconfig.CloudConfigVersion, prefixWorker)
	workerCloudConfig := BucketObjectState{
		Bucket: key.BucketName(customObject, accountID),
		Body:   workerBody,
		Key:    workerObjectName,
	}
	output[workerObjectName] = workerCloudConfig

	return output, nil
}
