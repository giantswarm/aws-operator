package s3object

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8scloudconfig/v_4_9_1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	kubernetesAPIHealthzVersion = "1c0cdf1ed5ee18fdf59063ecdd84bf3787f80fac"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var images v_4_9_1.Images
	var versions v_4_9_1.Versions
	{
		releaseVersion := customObject.Labels[label.ReleaseVersion]
		release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseVersion, metav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		findComponent := func(name string) (*v1alpha1.ReleaseSpecComponent, error) {
			for _, component := range release.Spec.Components {
				if component.Name == name {
					return &component, nil
				}
			}
			return nil, componentNotFoundError
		}

		{
			component, err := findComponent("kubernetes")
			if err != nil {
				return nil, err
			}
			versions.Kubernetes = component.Version
			images.Hyperkube = fmt.Sprintf("quay.io/giantswarm/hyperkube:%s", versions.Kubernetes)
		}

		{
			component, err := findComponent("etcd")
			if err != nil {
				return nil, err
			}
			images.Etcd = fmt.Sprintf("quay.io/giantswarm/etcd:%s", component.Version)
		}

		{
			component, err := findComponent("calico")
			if err != nil {
				return nil, err
			}
			versions.Calico = component.Version
			images.CalicoNode = fmt.Sprintf("quay.io/giantswarm/node:%s", component.Version)
			images.CalicoCNI = fmt.Sprintf("quay.io/giantswarm/cni:%s", component.Version)
			images.CalicoKubeControllers = fmt.Sprintf("quay.io/giantswarm/kube-controllers:%s", component.Version)
		}

		images.KubernetesAPIHealthz = fmt.Sprintf("quay.io/giantswarm/k8s-api-health:%s", kubernetesAPIHealthzVersion)
	}

	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	{
		g := &errgroup.Group{}

		g.Go(func() error {
			certs, err := r.certsSearcher.SearchCluster(key.ClusterID(customObject))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterCerts = certs

			return nil
		})

		g.Go(func() error {
			keys, err := r.randomKeysSearcher.SearchCluster(key.ClusterID(customObject))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterKeys = keys

			return nil
		})

		err = g.Wait()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	output := map[string]BucketObjectState{}
	{
		g := &errgroup.Group{}
		m := sync.Mutex{}

		data := cloudconfig.IgnitionTemplateData{
			CustomObject: customObject,
			ClusterCerts: clusterCerts,
			ClusterKeys:  clusterKeys,
			Images:       images,
			Versions:     versions,
		}
		g.Go(func() error {
			b, err := r.cloudConfig.NewMasterTemplate(ctx, data)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := key.BucketObjectName(customObject, key.KindMaster)
			output[k] = BucketObjectState{
				Bucket: key.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
				Body:   b,
				Key:    k,
			}
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			b, err := r.cloudConfig.NewWorkerTemplate(ctx, data)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := key.BucketObjectName(customObject, key.KindWorker)
			output[k] = BucketObjectState{
				Bucket: key.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
				Body:   b,
				Key:    k,
			}
			m.Unlock()

			return nil
		})

		err = g.Wait()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return output, nil
}
