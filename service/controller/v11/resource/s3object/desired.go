package s3object

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v11/cloudconfig"
	awsservicecontext "github.com/giantswarm/aws-operator/service/controller/v11/context/awsservice"
	cloudconfigcontext "github.com/giantswarm/aws-operator/service/controller/v11/context/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	output := map[string]BucketObjectState{}
	if err != nil {
		return output, microerror.Mask(err)
	}

	awsService, err := awsservicecontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	kmsKeyARN, err := awsService.GetKeyArn(clusterID)
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

	tlsAssets, err := r.encodeTLSAssets(ctx, certs, kmsKeyARN)
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterKeys, err := r.randomKeySearcher.SearchCluster(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	cc, err := cloudconfigcontext.FromContext(ctx)
	if err != nil {
		return output, microerror.Mask(err)
	}

	masterBody, err := cc.NewMasterTemplate(customObject, *tlsAssets, clusterKeys, kmsKeyARN)
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

	workerBody, err := cc.NewWorkerTemplate(customObject, *tlsAssets)
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
