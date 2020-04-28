package s3object

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	bn := key.BucketName(cr, cc.Status.TenantCluster.AWS.AccountID)

	// During deletion, it might happen that the encryption key got already
	// deleted. In such a case we do not have to do anything here anymore. The
	// desired state computation usually requires the encryption key to come up
	// with the deletion state, but in case it is gone we do not have to do
	// anything here anymore. The current implementation relies on the bucket
	// deletion of the s3bucket resource, which deletes all S3 objects and the
	// bucket itself.
	if cc.Status.TenantCluster.Encryption.Key == "" {
		if key.IsDeleted(cr) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key not available anymore")
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "encryption key not available yet")
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	paths, err := r.cloudConfig.NewPaths(ctx, obj)
	if hamaster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not computing current state", "reason", "control plane CR not available yet")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var s3Objects []*s3.PutObjectInput
	for _, p := range paths {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding S3 object %#q/%#q", bn, p))

		i := &s3.GetObjectInput{
			Bucket: aws.String(bn),
			Key:    aws.String(p),
		}

		o, err := cc.Client.TenantCluster.AWS.S3.GetObject(i)
		if IsBucketNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find S3 bucket %#q", bn))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if IsObjectNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find S3 object %#q/%#q", bn, p))
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		body, err = ioutil.ReadAll(o.Body)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found S3 object %#q/%#q", bn, p))

		s3Object := &s3.PutObjectInput{
			Key:           aws.String(p),
			Body:          strings.NewReader(string(body)),
			Bucket:        aws.String(bn),
			ContentLength: aws.Int64(int64(len(body))),
		}

		s3Objects = append(s3Objects, s3Object)
	}

	return s3Objects, nil
}
