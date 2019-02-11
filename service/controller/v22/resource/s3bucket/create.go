package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	createBucketsState, err := toBucketState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	g := &errgroup.Group{}

	for _, bucketInput := range createBucketsState {
		bucketInput := bucketInput

		g.Go(func() error {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 bucket %#q", bucketInput.Name))

			{
				i := &s3.CreateBucketInput{
					Bucket: aws.String(bucketInput.Name),
				}

				_, err = sc.AWSClient.S3.CreateBucket(i)
				if IsBucketAlreadyExists(err) {
					// Fall through.
				} else if IsBucketAlreadyOwnedByYou(err) {
					// Fall through.
				} else if err != nil {
					return microerror.Mask(err)
				}
			}

			if r.includeTags {
				i := &s3.PutBucketTaggingInput{
					Bucket: aws.String(bucketInput.Name),
					Tagging: &s3.Tagging{
						TagSet: r.getS3BucketTags(customObject),
					},
				}

				_, err = sc.AWSClient.S3.PutBucketTagging(i)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			if bucketInput.IsLoggingBucket {
				i := &s3.PutBucketAclInput{
					Bucket:       aws.String(key.TargetLogBucketName(customObject)),
					GrantReadACP: aws.String(key.LogDeliveryURI),
					GrantWrite:   aws.String(key.LogDeliveryURI),
				}

				_, err = sc.AWSClient.S3.PutBucketAcl(i)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			if bucketInput.IsLoggingBucket {
				i := &s3.PutBucketLifecycleConfigurationInput{
					Bucket: aws.String(key.TargetLogBucketName(customObject)),
					LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
						Rules: []*s3.LifecycleRule{
							{
								Expiration: &s3.LifecycleExpiration{
									Days: aws.Int64(int64(r.accessLogsExpiration)),
								},
								Filter: &s3.LifecycleRuleFilter{},
								ID:     aws.String(LifecycleLoggingBucketID),
								Status: aws.String("Enabled"),
							},
						},
					},
				}

				_, err = sc.AWSClient.S3.PutBucketLifecycleConfiguration(i)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			if bucketInput.IsLoggingEnabled {
				i := &s3.PutBucketLoggingInput{
					Bucket: aws.String(bucketInput.Name),
					BucketLoggingStatus: &s3.BucketLoggingStatus{
						LoggingEnabled: &s3.LoggingEnabled{
							TargetBucket: aws.String(key.TargetLogBucketName(customObject)),
							TargetPrefix: aws.String(bucketInput.Name + "/"),
						},
					},
				}

				_, err = sc.AWSClient.S3.PutBucketLogging(i)
				if err != nil {
					return microerror.Mask(err)
				}
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 bucket %#q", bucketInput.Name))

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentBuckets, err := toBucketState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredBuckets, err := toBucketState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createState []BucketState
	for _, bucket := range desiredBuckets {
		if !containsBucketState(bucket.Name, currentBuckets) {
			createState = append(createState, bucket)
		}
	}

	return createState, nil
}
