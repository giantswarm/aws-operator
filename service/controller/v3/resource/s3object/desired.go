package s3object

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v3/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v3/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	output := map[string]BucketObjectState{}
	if err != nil {
		return output, microerror.Mask(err)
	}

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	kmsArn, err := r.awsService.GetKeyArn(clusterID)
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

	randomKeys, err := r.randomKeyWatcher.SearchKeys(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	tlsAssets, err := r.encodeTLSAssets(certs, kmsArn)
	if err != nil {
		return output, microerror.Mask(err)
	}

	randomKeyAssets, err := r.encodeKeyAssets(randomKeys, kmsArn)
	if err != nil {
		return output, microerror.Mask(err)
	}

	masterBody, err := r.cloudConfig.NewMasterTemplate(customObject, *tlsAssets, *randomKeyAssets)
	if err != nil {
		return output, microerror.Mask(err)
	}

	masterObjectName := key.BucketObjectName(cloudconfig.MasterCloudConfigVersion, prefixMaster)
	masterCloudConfig := BucketObjectState{
		Bucket: key.BucketName(customObject, accountID),
		Body:   masterBody,
		Key:    masterObjectName,
	}
	output[masterObjectName] = masterCloudConfig

	workerBody, err := r.cloudConfig.NewWorkerTemplate(customObject, *tlsAssets)
	if err != nil {
		return output, microerror.Mask(err)
	}

	workerObjectName := key.BucketObjectName(cloudconfig.WorkerCloudConfigVersion, prefixWorker)
	workerCloudConfig := BucketObjectState{
		Bucket: key.BucketName(customObject, accountID),
		Body:   workerBody,
		Key:    workerObjectName,
	}
	output[workerObjectName] = workerCloudConfig

	return output, nil
}
