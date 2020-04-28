package s3object

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
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

	releaseVersion := key.ReleaseVersion(cr)

	var cluster infrastructurev1alpha2.AWSCluster
	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	var release *v1alpha1.Release
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

		g.Go(func() error {
			releaseCR, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(key.ReleaseName(releaseVersion), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			release = releaseCR

			return nil
		})

		err = g.Wait()
		if certs.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "certificate secrets are not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if randomkeys.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "random key secrets are not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var images k8scloudconfig.Images
	{
		v, err := k8scloudconfig.ExtractComponentVersions(release.Spec.Components)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		v.Kubectl = key.KubectlVersion
		v.KubernetesAPIHealthz = key.KubernetesAPIHealthzVersion
		v.KubernetesNetworkSetupDocker = key.K8sSetupNetworkEnvironment
		images = k8scloudconfig.BuildImages(r.registryDomain, v)
	}

	var s3Objects []*s3.PutObjectInput

	cp, isControlPlaneCR := obj.(*infrastructurev1alpha2.AWSControlPlane)
	if isControlPlaneCR {
		// tccpn render

		// Fetch the replicas number.
		g8scp, err := r.g8sClient.InfrastructureV1alpha2().G8sControlPlanes(cp.Namespace).Get(cp.Name, metav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
		// get replicas num
		numReplicas := g8scp.Spec.Replicas

		for i := 0; i < numReplicas; i++ {
			body, err := r.cloudConfig.RenderTCCPN(ctx, cluster, clusterCerts, clusterKeys, images, i, r.labelsFunc(cr))
			if err != nil {
				return nil, microerror.Mask(err)
			}

			s3Object :=
				&s3.PutObjectInput{
					Key:           aws.String(key.ControlPlaneMasterIgnitionPath(r.pathFunc(cr), i)),
					Body:          strings.NewReader(string(body)),
					Bucket:        aws.String(key.BucketName(cr, cc.Status.TenantCluster.AWS.AccountID)),
					ContentLength: aws.Int64(int64(len(body))),
				}
			s3Objects = append(s3Objects, s3Object)
		}
	} else {
		// tcnp render
		body, err := r.cloudConfig.RenderTCNP(ctx, cluster, clusterCerts, clusterKeys, images, r.labelsFunc(cr))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		s3Object :=
			&s3.PutObjectInput{
				Key:           aws.String(r.pathFunc(cr)),
				Body:          strings.NewReader(string(body)),
				Bucket:        aws.String(key.BucketName(cr, cc.Status.TenantCluster.AWS.AccountID)),
				ContentLength: aws.Int64(int64(len(body))),
			}
		s3Objects = append(s3Objects, s3Object)

	}

	return s3Objects, nil
}
