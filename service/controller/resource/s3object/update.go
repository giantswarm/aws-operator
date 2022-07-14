package s3object

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/crud"

	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	s3Objects, err := toS3Objects(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(s3Objects) != 0 {
		for _, s3Object := range s3Objects {
			r.logger.Debugf(ctx, "updating S3 object %#q", *s3Object.Key)

			_, err = cc.Client.TenantCluster.AWS.S3.PutObject(s3Object)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "updated S3 object %#q", *s3Object.Key)
		}
	} else {
		r.logger.Debugf(ctx, "did not update any S3 object")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Objects, err := toS3Objects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Objects, err := toS3Objects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(currentS3Objects) == 0 {
		// In case there is no current state we need to create first and cannot
		// update.
		return nil, nil
	}

	// We do a poor man's comparison here to figure out if we have to deal with a
	// change between current and desired state. The first and most straight
	// forward thing to do at this point is to simply check how many S3 Objects we
	// have. When current and desired state do not have the same number of items,
	// we simply apply all of the desired state without being smarter about it.
	var num int
	{
		if len(currentS3Objects) != len(desiredS3Objects) {
			return desiredS3Objects, nil
		}

		num = len(currentS3Objects)
	}

	// The passed resource state defines the actual Cloud Config content as
	// io.ReadSeaker. In order to compare the current and desired state we need to
	// read and re-apply the byte stream once we read it. Otherwise we would flush
	// content and it would not be available anymore for create or update calls.
	// Note that we apply the same primitive comparison here as described above.
	// In case one item of current state does equal the desired item at the same
	// position in the list, we simply apply all of the desired state without
	// being smarter about it.
	for i := 0; i < num; i++ {
		currentS3Object := currentS3Objects[i]
		desiredS3Object := desiredS3Objects[i]

		var c []byte
		{
			c, err = ioutil.ReadAll(currentS3Object.Body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			currentS3Object.Body = strings.NewReader(string(c))
		}

		var d []byte
		{
			d, err = ioutil.ReadAll(desiredS3Object.Body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			desiredS3Object.Body = strings.NewReader(string(d))
		}

		if !bytes.Equal(c, d) {
			return desiredS3Objects, nil
		}
	}

	return nil, nil
}
