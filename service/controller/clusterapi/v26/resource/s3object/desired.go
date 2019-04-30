package s3object

import (
	"context"
	"sync"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/legacykey"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := legacykey.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	{
		g := &errgroup.Group{}

		g.Go(func() error {
			certs, err := r.certsSearcher.SearchCluster(legacykey.ClusterID(customObject))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterCerts = certs

			return nil
		})

		g.Go(func() error {
			keys, err := r.randomKeysSearcher.SearchCluster(legacykey.ClusterID(customObject))
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

		g.Go(func() error {
			b, err := r.cloudConfig.NewMasterTemplate(ctx, customObject, clusterCerts, clusterKeys)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := legacykey.BucketObjectName(customObject, legacykey.KindMaster)
			output[k] = BucketObjectState{
				Bucket: legacykey.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
				Body:   b,
				Key:    k,
			}
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			b, err := r.cloudConfig.NewWorkerTemplate(ctx, customObject, clusterCerts)
			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			k := legacykey.BucketObjectName(customObject, legacykey.KindWorker)
			output[k] = BucketObjectState{
				Bucket: legacykey.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
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
