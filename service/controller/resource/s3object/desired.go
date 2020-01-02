package s3object

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	var cluster infrastructurev1alpha2.AWSCluster
	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	{
		g := &errgroup.Group{}

		g.Go(func() error {
			m, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters(cr.GetNamespace()).Get(key.ClusterID(cr), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			cluster = *m

			return nil
		})

		g.Go(func() error {
			certs, err := r.certsSearcher.SearchCluster(key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterCerts = certs

			return nil
		})

		g.Go(func() error {
			keys, err := r.randomKeysSearcher.SearchCluster(key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterKeys = keys

			return nil
		})

		err = g.Wait()
		if certs.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "certificate secrets are not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if randomkeys.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "random key secrets are not yet available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	body, err := r.cloudConfig.Render(ctx, cluster, clusterCerts, clusterKeys, r.labelsFunc(cr))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s3Object := &s3.PutObjectInput{
		Key:           aws.String(r.pathFunc(cr)),
		Body:          strings.NewReader(string(body)),
		Bucket:        aws.String(key.BucketName(cr, cc.Status.TenantCluster.AWS.AccountID)),
		ContentLength: aws.Int64(int64(len(body))),
	}

	return s3Object, nil
}
