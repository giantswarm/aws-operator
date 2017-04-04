package common

import (
	"encoding/json"
	"sync"

	"github.com/giantswarm/awstpr"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	k8sutil "github.com/giantswarm/aws-operator/client/k8s"
)

const (
	ClusterListAPIEndpoint  string = "/apis/cluster.giantswarm.io/v1/awses"
	ClusterWatchAPIEndpoint string = "/apis/cluster.giantswarm.io/v1/watch/awses"
)

type Config struct {
	AwsConfig awsutil.Config
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Service struct {
	// Dependencies.
	AwsConfig awsutil.Config
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Internals.
	BootOnce sync.Once
}

func (s *Service) NewClusterListWatch() *cache.ListWatch {
	client := s.K8sClient.Core().RESTClient()

	listWatch := &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			req := client.Get().AbsPath(ClusterListAPIEndpoint)
			b, err := req.DoRaw()
			if err != nil {
				return nil, err
			}

			var c awstpr.List
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
