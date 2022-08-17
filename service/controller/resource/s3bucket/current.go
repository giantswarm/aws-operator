package s3bucket

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketStateNames := []string{
		key.TargetLogBucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
		key.BucketName(&cr, cc.Status.TenantCluster.AWS.AccountID),
	}

	var currentBucketState []BucketState
	{
		r.logger.Debugf(ctx, "finding the S3 buckets")

		g := &errgroup.Group{}
		m := sync.Mutex{}

		for _, inputBucketName := range bucketStateNames {
			bucketName := inputBucketName

			g.Go(func() error {
				inputBucket := BucketState{
					Name: bucketName,
				}

				// TODO this check should not be done here. Here we only fetch the
				// current state. We have to make a request anyway so fetching what we
				// want and handling the not found errors as usual should be the way to
				// go.
				//
				//
				//     https://github.com/giantswarm/giantswarm/issues/5246
				//
				isCreated, err := r.isBucketCreated(ctx, bucketName)
				if err != nil {
					return microerror.Mask(err)
				}
				if !isCreated {
					return nil
				}

				r.logger.Debugf(ctx, "finding the S3 bucket %#q", bucketName)

				lc, err := r.getLoggingConfiguration(ctx, bucketName)
				if err != nil {
					return microerror.Mask(err)
				}

				m.Lock()
				inputBucket.IsLoggingBucket = isLoggingBucket(bucketName, lc)
				inputBucket.IsLoggingEnabled = isLoggingEnabled(lc)
				currentBucketState = append(currentBucketState, inputBucket)
				m.Unlock()

				r.logger.Debugf(ctx, "found the S3 bucket %#q", bucketName)

				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found the S3 buckets")
	}

	return currentBucketState, nil
}

func (r *Resource) isBucketCreated(ctx context.Context, name string) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	headInput := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}
	_, err = cc.Client.TenantCluster.AWS.S3.HeadBucket(headInput)
	if IsBucketNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (r *Resource) getLoggingConfiguration(ctx context.Context, name string) (*s3.GetBucketLoggingOutput, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketLoggingInput := &s3.GetBucketLoggingInput{
		Bucket: aws.String(name),
	}
	bucketLoggingOutput, err := cc.Client.TenantCluster.AWS.S3.GetBucketLogging(bucketLoggingInput)
	if err != nil {
		return bucketLoggingOutput, microerror.Mask(err)
	}

	return bucketLoggingOutput, nil
}

func isLoggingEnabled(lc *s3.GetBucketLoggingOutput) bool {
	return lc.LoggingEnabled != nil
}

func isLoggingBucket(name string, lc *s3.GetBucketLoggingOutput) bool {
	if lc.LoggingEnabled != nil {
		if *lc.LoggingEnabled.TargetBucket == name {
			return true
		}
	}

	return false
}
