package statusresource

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type CollectorSetConfig struct {
	Logger  micrologger.Logger
	Watcher func(opts metav1.ListOptions) (watch.Interface, error)
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the iniitialization and prevents some weird import mess so we do not
// have to alias packages. There is also the benefit of the helper type kept
// private so we do not need to expose this magic.
type CollectorSet struct {
	*collector.Set
}

func NewCollectorSet(config CollectorSetConfig) (*CollectorSet, error) {
	var err error

	var legacyStatusCollector *LegacyStatusCollector
	{
		c := LegacyStatusCollectorConfig{
			Logger:  config.Logger,
			Watcher: config.Watcher,
		}

		legacyStatusCollector, err = NewLegacyStatusCollector(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				legacyStatusCollector,
			},
			Logger: config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &CollectorSet{
		Set: collectorSet,
	}

	return s, nil
}
