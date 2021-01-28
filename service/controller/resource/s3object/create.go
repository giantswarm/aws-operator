package s3object

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	s3Objects, err := toS3Objects(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(s3Objects) != 0 {
		for _, s3Object := range s3Objects {
			r.logger.Debugf(ctx, "creating S3 object %#q", *s3Object.Key)

			_, err = cc.Client.TenantCluster.AWS.S3.PutObject(s3Object)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "created S3 object %#q", *s3Object.Key)
		}
	} else {
		r.logger.Debugf(ctx, "did not create any S3 object")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Objects, err := toS3Objects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Objects, err := toS3Objects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createState []*s3.PutObjectInput
	if len(currentS3Objects) == 0 {
		createState = desiredS3Objects
	}

	return createState, nil
}
