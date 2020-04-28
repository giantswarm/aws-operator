package s3object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	s3Objects, err := toS3ObjectArray(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(s3Objects) != 0 {
		for i := 0; i < len(s3Objects); i++ {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object %#q", s3Objects[i].Key))

			_, err = cc.Client.TenantCluster.AWS.S3.PutObject(s3Objects[i])
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 object %#q", s3Objects[i].Key))
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not create S3 object %#q", r.pathFunc(cr)))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Objects, err := toS3ObjectArray(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Objects, err := toS3ObjectArray(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createState []*s3.PutObjectInput
	if len(currentS3Objects) == 0 {
		createState = desiredS3Objects
	}

	return createState, nil
}
