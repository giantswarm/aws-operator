package s3objectv2

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	output := BucketObjectState{}
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return output, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for S3 objects")

	workersObjectName := keyv2.BucketObjectName(customObject, prefixWorker)

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	bucketName := keyv2.BucketName(customObject, accountID)

	input := &s3.GetObjectInput{
		Key:    aws.String(workersObjectName),
		Bucket: aws.String(bucketName),
	}
	result, err := r.awsClients.S3.GetObject(input)

	if IsObjectNotFound(err) || IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find the S3 objects")
		// fall through

	} else if err != nil {

		return output, microerror.Mask(err)

	} else {

		r.logger.LogCtx(ctx, "debug", "found the S3 objects")
		output.WorkerCloudConfig.Key = workersObjectName
		output.WorkerCloudConfig.Bucket = bucketName

		buf := new(bytes.Buffer)
		buf.ReadFrom(result.Body)
		body := buf.String()

		output.WorkerCloudConfig.Body = body
	}

	return output, nil
}
