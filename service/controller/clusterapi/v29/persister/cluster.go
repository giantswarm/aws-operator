
package persister

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type Config struct {
	CMAClient clientset.Interface
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	NetworkRange net.IPNet
}

type Collector struct {
	cmaClient clientset.Interface
	g8sClient versioned.Interface
	logger    micrologger.Logger

	networkRange net.IPNet
}

func New(config Config) (*Collector, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if reflect.DeepEqual(config.NetworkRange, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkRange must not be empty", config)
	}

	c := &Collector{
		cmaClient: config.CMAClient,
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		networkRange: config.NetworkRange,
	}

	return c, nil
}

  func (c *Cluster) NewPersistFunc(ctx context.Contex, obj interface{}) network.Persister {
	return func(ctx context.Context, subnet net.IPNet) error {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status to persist network allocation")

		var providerStatus g8sv1alpha1.AWSClusterStatus
		{
			err := json.Unmarshal(cr.Status.ProviderStatus.Raw, &providerStatus)
			if err != nil {
				return microerror.Mask(err)
			}

			providerStatus.Provider.Network.CIDR = subnet.String()
		}

		{
			b, err := json.Marshal(providerStatus)
			if err != nil {
				return microerror.Mask(err)
			}

			cr.Status.ProviderStatus.Raw = b
		}

		{
			_, err := r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).UpdateStatus(&cr)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated CR status to persist network allocation")

		return nil
	}
  }
