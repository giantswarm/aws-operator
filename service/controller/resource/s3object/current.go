package s3object

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
	"github.com/giantswarm/aws-operator/v14/service/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/v14/service/internal/encrypter/kms"
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

	_, err = r.encrypter.EncryptionKey(ctx, key.ClusterID(cr))
	if kms.IsKeyNotFound(err) {
		r.logger.Debugf(ctx, "canceling resource", "reason", "encryption key not available yet")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if kms.IsKeyScheduledForDeletion(err) {
		r.logger.Debugf(ctx, "canceling resource", "reason", "encryption key not available anymore")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	paths, err := r.cloudConfig.NewPaths(ctx, obj)
	if cloudconfig.IsNotFound(err) {
		r.logger.Debugf(ctx, "not computing current state", "reason", "control plane CR not available yet")
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	var s3Objects []*s3.PutObjectInput
	for _, p := range paths {
		r.logger.Debugf(ctx, "finding S3 object %#q", fmt.Sprintf("%s/%s", bn, p))

		i := &s3.GetObjectInput{
			Bucket: aws.String(bn),
			Key:    aws.String(p),
		}

		o, err := cc.Client.TenantCluster.AWS.S3.GetObject(i)
		if IsBucketNotFound(err) {
			r.logger.Debugf(ctx, "not computing current state", "reason", fmt.Sprintf("did not find S3 bucket %#q", bn))
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if IsObjectNotFound(err) {
			r.logger.Debugf(ctx, "did not find S3 object %#q", fmt.Sprintf("%s/%s", bn, p))
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		body, err := ioutil.ReadAll(o.Body)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found S3 object %#q", fmt.Sprintf("%s/%s", bn, p))

		s3Object := &s3.PutObjectInput{
			Key:           aws.String(p),
			Body:          strings.NewReader(string(body)),
			Bucket:        aws.String(bn),
			ContentLength: aws.Int64(int64(len(body))),
		}

		s3Objects = append(s3Objects, s3Object)
	}

	// We want to prevent Cloud Formation stacks from being created without the
	// Cloud Config being uploaded to S3. The TCCPN and TCNP handlers check this
	// value and cancel in case the S3 Object is not yet uploaded.
	cc.Status.TenantCluster.S3Object.Uploaded = true

	return s3Objects, nil
}
