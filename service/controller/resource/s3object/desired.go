package s3object

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudconfig"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	paths, err := r.cloudConfig.NewPaths(ctx, obj)
	if cloudconfig.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state", "reason", "control plane CR not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	templates, err := r.cloudConfig.NewTemplates(ctx, obj)
	if cloudconfig.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state", "reason", "control plane CR not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if cloudconfig.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing desired state", "reason", "secrets are not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var num int
	{
		if len(paths) != len(templates) {
			return nil, microerror.Maskf(executionFailedError, "cloud config implementation produced invalid result")
		}

		num = len(paths)
	}

	var s3Objects []*s3.PutObjectInput
	for i := 0; i < num; i++ {
		p := paths[i]
		t := templates[i]

		s3Object := &s3.PutObjectInput{
			Key:           aws.String(p),
			Body:          strings.NewReader(string(t)),
			Bucket:        aws.String(key.BucketName(cr, cc.Status.TenantCluster.AWS.AccountID)),
			ContentLength: aws.Int64(int64(len(t))),
		}

		s3Objects = append(s3Objects, s3Object)
	}

	return s3Objects, nil
}
