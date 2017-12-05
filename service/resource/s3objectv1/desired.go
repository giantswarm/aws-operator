package s3objectv1

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	output := BucketObjectState{}
	if err != nil {
		return output, microerror.Mask(err)
	}

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	certs, err := r.certWatcher.SearchCerts(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	kmsArn, err := r.awsService.GetKeyArn(clusterID)
	if err != nil {
		return output, microerror.Mask(err)
	}

	tlsAssets, err := r.encodeTLSAssets(certs, kmsArn)
	if err != nil {
		return output, microerror.Mask(err)
	}

	body, err := r.cloudConfig.NewWorkerTemplate(customObject, *tlsAssets)
	if err != nil {
		return output, microerror.Mask(err)
	}

	output.WorkerCloudConfig = BucketObjectInstance{
		Bucket: key.BucketName(customObject, accountID),
		Body:   body,
		Key:    key.BucketObjectName(customObject, prefixWorker),
	}

	return output, nil
}
