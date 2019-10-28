package s3object

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	s3Object, err := toS3Object(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if s3Object != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating S3 object %#q", *s3Object.Key))

		_, err = cc.Client.TenantCluster.AWS.S3.PutObject(s3Object)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated S3 object %#q", *s3Object.Key))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not update S3 object %#q", r.pathFunc(cr)))
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentS3Object, err := toS3Object(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredS3Object, err := toS3Object(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// The passed resource state defines the actual Cloud Config content as
	// io.ReadSeaker. In order to compare the current and desired state we need to
	// read and re-apply the byte stream once we read it. Otherwise we would flush
	// content and it would not be available anymore for create or update calls.
	var updateState *s3.PutObjectInput
	{
		var c []byte
		if currentS3Object != nil {
			c, err = ioutil.ReadAll(currentS3Object.Body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			currentS3Object.Body = strings.NewReader(string(c))
		} else {
			// In case there is no current state we need to create first and cannot
			// update.
			return nil, nil
		}

		var d []byte
		if desiredS3Object != nil {
			d, err = ioutil.ReadAll(desiredS3Object.Body)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			desiredS3Object.Body = strings.NewReader(string(d))
		}

		if !bytes.Equal(c, d) {
			updateState = desiredS3Object
		}
	}

	return updateState, nil
}
