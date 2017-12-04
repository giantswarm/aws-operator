package s3workercloudconfigv1

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	s3PutInput, err := toPutObjectInput(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.awsClients.S3.PutObject(&s3PutInput)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "creating S3 worker's cloudconfig: created")

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	desiredObjectState, err := toBucketObject(desiredState)
	if err != nil {
		return s3.PutObjectInput{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "finding out if worker's cloudconfig object should be created")

	createState := s3.PutObjectInput{
		Key: aws.String(""),
	}

	if desiredObjectState.Key != "" {
		createState.Key = aws.String(desiredObjectState.Key)
		createState.Body = strings.NewReader(desiredObjectState.Body)
		createState.Bucket = aws.String(desiredObjectState.Bucket)
		createState.ContentLength = aws.Int64(int64(len(desiredObjectState.Body)))
	}

	return createState, nil
}
