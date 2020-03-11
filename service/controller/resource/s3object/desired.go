package s3object

import (
	"context"
	"fmt"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	gscerts "github.com/giantswarm/certs"
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
	kubectlVersion              = "9ccdc9dc55a01b1fde2aea73901d0a699909c9cd" // 1.15.5
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
	{
		releaseVersion := customObject.Labels[label.ReleaseVersion]
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseName, metav1.GetOptions{})
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
			images.Hyperkube = fmt.Sprintf("%s/giantswarm/hyperkube:v%s", r.registryDomain, component.Version)
		}

		{
			component, err := findComponent("etcd")
			if err != nil {
				return nil, err
			}
			images.Etcd = fmt.Sprintf("%s/giantswarm/etcd:v%s", r.registryDomain, component.Version)
		}

		{
			component, err := findComponent("calico")
			if err != nil {
				return nil, err
			}
			images.CalicoNode = fmt.Sprintf("%s/giantswarm/node:v%s", r.registryDomain, component.Version)
			images.CalicoCNI = fmt.Sprintf("%s/giantswarm/cni:v%s", r.registryDomain, component.Version)
			images.CalicoKubeControllers = fmt.Sprintf("%s/giantswarm/kube-controllers:v%s", r.registryDomain, component.Version)
		}

		images.Kubectl = fmt.Sprintf("%s/giantswarm/docker-kubectl:%s", r.registryDomain, kubectlVersion)
		images.KubernetesAPIHealthz = fmt.Sprintf("%s/giantswarm/k8s-api-health:%s", r.registryDomain, kubernetesAPIHealthzVersion)
	}

	var clusterCerts gscerts.Cluster
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
