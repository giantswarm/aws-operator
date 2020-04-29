package s3object

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"sync"

	gscerts "github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	var images k8scloudconfig.Images
	{
		versions, err := k8scloudconfig.ExtractComponentVersions(cc.Spec.TenantCluster.Release.Spec.Components)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		defaultVersions := key.DefaultVersions()
		versions.Kubectl = defaultVersions.Kubectl
		versions.KubernetesAPIHealthz = defaultVersions.KubernetesAPIHealthz
		versions.KubernetesNetworkSetupDocker = defaultVersions.KubernetesNetworkSetupDocker
		images = k8scloudconfig.BuildImages(r.registryDomain, versions)
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
			ignition, err := r.cloudConfig.NewMasterTemplate(ctx, data)
			if err != nil {
				return microerror.Mask(err)
			}

			decrypted, err := r.cloudConfig.DecryptTemplate(ctx, ignition)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := key.BucketObjectName(key.KindMaster)
			object := BucketObjectState{
				Bucket: key.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
				Body:   ignition,
				Key:    k,
				Hash:   hashIgnition(decrypted),
			}
			output[k] = object
			cc.Spec.TenantCluster.MasterInstance.IgnitionHash = object.Hash
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			ignition, err := r.cloudConfig.NewWorkerTemplate(ctx, data)
			if err != nil {
				return microerror.Mask(err)
			}

			decrypted, err := r.cloudConfig.DecryptTemplate(ctx, ignition)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := key.BucketObjectName(key.KindWorker)
			object := BucketObjectState{
				Bucket: key.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
				Body:   ignition,
				Key:    k,
				Hash:   hashIgnition(decrypted),
			}
			output[k] = object
			cc.Spec.TenantCluster.WorkerInstance.IgnitionHash = object.Hash
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

// hashIgnition returns a hash value representing the given ignition.
func hashIgnition(encoded string) string {
	rawSum := sha512.Sum512([]byte(encoded))
	sum := rawSum[:]
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)
	return string(encodedSum)
}
