package locker

import (
	"context"
	"fmt"
	"time"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/project"
)

const (
	timeLockName = "timelock.giantswarm.io/until"
)

var (
	timeLockOwner = fmt.Sprintf("%s-%s", timeLockName, project.Name())
)

type TimeLockConfig struct {
	Logger    micrologger.Logger
	K8sClient client.Client

	ClusterID          string
	ClusterCRNamespace string
	TTL                time.Duration
}

type TimeLock struct {
	logger    micrologger.Logger
	k8sClient client.Client

	clusterID          string
	clusterCRNamespace string
	ttl                time.Duration
}

// NewTimeLock implements a distributed time lock mechanism mainly used for node auto repair pause period
// You can inspect the lock annotations on the AWSCluster CR.
// The lock is unique to each cluster ID.
//     $ kubectl get awscluster $CLUSTER_ID --watch | jq '.metadata.annotations'
//     "timelock.giantswarm.io/aws-operator@6.7.0": "Mon Jan 2 15:04:05 MST 2006"
//
func NewTimeLock(config TimeLockConfig) (*TimeLock, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.TTL == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.TTL must not be zero", config)
	}

	d := &TimeLock{
		logger:    config.Logger,
		k8sClient: config.K8sClient,

		clusterID:          config.ClusterID,
		clusterCRNamespace: config.ClusterCRNamespace,
		ttl:                config.TTL,
	}

	return d, nil
}

func (t *TimeLock) Lock(ctx context.Context) error {
	locked, err := t.isLocked(ctx)
	if err != nil {
		return err
	}

	if locked {
		// fail since lock is already acquired
		return microerror.Maskf(alreadyExistsError, fmt.Sprintf("time lock for cluster %s already exists", t.clusterID))
	}

	err = t.createLock(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (t *TimeLock) isLocked(ctx context.Context) (bool, error) {
	var err error
	isLocked := false

	var awsCluster infrastructurev1alpha2.AWSCluster

	err = t.k8sClient.Get(ctx, types.NamespacedName{Namespace: t.clusterCRNamespace, Name: t.clusterID}, &awsCluster)
	if err != nil {
		return false, microerror.Mask(err)
	}

	timeStamp, ok := awsCluster.Annotations[timeLockOwner]
	if ok {
		ts, err := time.Parse(time.UnixDate, timeStamp)
		if err != nil {
			return false, microerror.Mask(err)
		}
		// check fi the lock is expired
		if time.Now().Before(ts) {
			isLocked = true
		}
	}
	return isLocked, nil
}

func (t *TimeLock) createLock(ctx context.Context) error {
	var awsCluster infrastructurev1alpha2.AWSCluster

	err := t.k8sClient.Get(ctx, types.NamespacedName{Namespace: t.clusterCRNamespace, Name: t.clusterID}, &awsCluster)
	if err != nil {
		return microerror.Mask(err)
	}
	// add lock timestamp
	awsCluster.Annotations[timeLockOwner] = time.Now().Add(t.ttl).Format(time.UnixDate)

	err = t.k8sClient.Update(ctx, &awsCluster)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (t *TimeLock) clearLock(ctx context.Context) error {
	var awsCluster infrastructurev1alpha2.AWSCluster

	err := t.k8sClient.Get(ctx, types.NamespacedName{Namespace: t.clusterCRNamespace, Name: t.clusterID}, &awsCluster)
	if err != nil {
		return microerror.Mask(err)
	}

	updated := false

	if _, ok := awsCluster.Annotations[timeLockOwner]; ok {
		// delete lock from annotations
		delete(awsCluster.Annotations, timeLockOwner)
		updated = true
	}

	if updated {
		err = t.k8sClient.Update(ctx, &awsCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}
