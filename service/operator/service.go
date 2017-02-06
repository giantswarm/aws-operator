package operator

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/giantswarm/clusterspec"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
)

const (
	ClusterListAPIEndpoint  = "/apis/giantswarm.io/v1/clusters"
	ClusterWatchAPIEndpoint = "/apis/giantswarm.io/v1/watch/clusters"
	// Period or re-synchronizing the list of objects in k8s watcher. 0 means that re-sync will be
	// delayed as long as possible, until the watch will be closed or timed out.
	resyncPeriod = 0
)

// Config represents the configuration used to create a version service.
type Config struct {
	// Dependencies.
	K8sclient kubernetes.Interface
	Logger    micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new version service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:    nil,
		K8sclient: nil,
	}
}

// New creates a new configured version service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	newService := &Service{
		// Dependencies.
		k8sclient: config.K8sclient,
		logger:    config.Logger,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service implements the version service interface.
type Service struct {
	// Dependencies.
	logger    micrologger.Logger
	k8sclient kubernetes.Interface

	// Internals.
	bootOnce sync.Once
}

type Event struct {
	Type   string
	Object *clusterspec.Cluster
}

func (s *Service) newClusterListWatch() *cache.ListWatch {
	client := s.k8sclient.Core().RESTClient()

	listWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			req := client.Get().AbsPath(ClusterListAPIEndpoint)
			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}

			var c clusterspec.ClusterList
			if err := json.Unmarshal(b, &c); err != nil {
				return nil, err
			}

			return &c, nil
		},

		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			req := client.Get().AbsPath(ClusterWatchAPIEndpoint)
			stream, err := req.Stream()
			if err != nil {
				return nil, err
			}

			watcher := watch.NewStreamWatcher(&k8sutil.ClusterDecoder{
				Stream: stream,
			})

			return watcher, nil
		},
	}

	return listWatch
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		if err := s.createTPR(); err != nil {
			panic(err)
		}
		s.logger.Log("info", "successfully created third-party resource")

		_, clusterInformer := cache.NewInformer(
			s.newClusterListWatch(),
			&clusterspec.Cluster{},
			resyncPeriod,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					cluster := obj.(*clusterspec.Cluster)
					s.logger.Log("info", fmt.Sprintf("cluster '%v' added", cluster.Name))

					if err := s.createClusterNamespace(*cluster); err != nil {
						s.logger.Log("error", "could not create cluster namespace:", err)
					}
				},
				DeleteFunc: func(obj interface{}) {
					cluster := obj.(*clusterspec.Cluster)
					s.logger.Log("info", fmt.Sprintf("cluster '%v' deleted", cluster.Name))

					if err := s.deleteClusterNamespace(*cluster); err != nil {
						s.logger.Log("error", "could not delete cluster namespace:", err)
					}
				},
			},
		)

		s.logger.Log("info", "starting watch")

		// Cluster informer lifecycle can be interrupted by putting a value into a "stop channel".
		// We aren't currently using that functionality, so we are passing a nil here.
		clusterInformer.Run(nil)
	})
}
