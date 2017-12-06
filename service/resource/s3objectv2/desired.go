package s3objectv2

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv1.ToCustomObject(obj)
	output := BucketObjectState{}
	if err != nil {
		return output, microerror.Mask(err)
	}

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	clusterID := keyv1.ClusterID(customObject)
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
		Bucket: keyv1.BucketName(customObject, accountID),
		Body:   body,
		Key:    keyv1.BucketObjectName(customObject, prefixWorker),
	}

	return output, nil
}
