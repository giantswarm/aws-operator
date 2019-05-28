package statusresource

import (
	"sync"
	"time"

	providerv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	legacyStatusCollectorDescription *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName("statusresource", "cluster", "status"),
		"Cluster status condition as provided by the CR status.",
		[]string{
			"cluster_id",
			"status",
		},
		nil,
	)
	legacyCreationSecondsCollectorDescription *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName("statusresource", "cluster", "creation_seconds"),
		"Cluster creation duration.",
		[]string{
			"cluster_id",
		},
		nil,
	)
)

type LegacyStatusCollectorConfig struct {
	Logger  micrologger.Logger
	Watcher func(opts metav1.ListOptions) (watch.Interface, error)
}

type LegacyStatusCollector struct {
	logger  micrologger.Logger
	watcher func(opts metav1.ListOptions) (watch.Interface, error)

	bootOnce sync.Once
}

func NewLegacyStatusCollector(config LegacyStatusCollectorConfig) (*LegacyStatusCollector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Watcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Watcher must not be empty", config)
	}

	c := &LegacyStatusCollector{
		logger:  config.Logger,
		watcher: config.Watcher,

		bootOnce: sync.Once{},
	}

	return c, nil
}

func (c *LegacyStatusCollector) Collect(ch chan<- prometheus.Metric) error {
	watcher, err := c.watcher(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				continue
			}

			m, err := meta.Accessor(event.Object)
			if err != nil {
				return microerror.Mask(err)
			}
			p, ok := event.Object.(Provider)
			if !ok {
				panic("asserting Provider interface failed")
			}

			if p.ClusterStatus().HasCreatedCondition() {
				t1 := timeForCreatedCondition(p)
				t2 := m.GetCreationTimestamp().Time

				ch <- prometheus.MustNewConstMetric(
					legacyCreationSecondsCollectorDescription,
					prometheus.GaugeValue,
					t1.Sub(t2).Seconds(),
					m.GetName(),
				)
			}

			{
				ch <- prometheus.MustNewConstMetric(
					legacyStatusCollectorDescription,
					prometheus.GaugeValue,
					float64(boolToInt(p.ClusterStatus().HasCreatingCondition())),
					m.GetName(),
					providerv1alpha1.StatusClusterTypeCreating,
				)
				ch <- prometheus.MustNewConstMetric(
					legacyStatusCollectorDescription,
					prometheus.GaugeValue,
					float64(boolToInt(p.ClusterStatus().HasCreatedCondition())),
					m.GetName(),
					providerv1alpha1.StatusClusterTypeCreated,
				)
				ch <- prometheus.MustNewConstMetric(
					legacyStatusCollectorDescription,
					prometheus.GaugeValue,
					float64(boolToInt(p.ClusterStatus().HasUpdatingCondition())),
					m.GetName(),
					providerv1alpha1.StatusClusterTypeUpdating,
				)
				ch <- prometheus.MustNewConstMetric(
					legacyStatusCollectorDescription,
					prometheus.GaugeValue,
					float64(boolToInt(p.ClusterStatus().HasUpdatedCondition())),
					m.GetName(),
					providerv1alpha1.StatusClusterTypeUpdated,
				)
				ch <- prometheus.MustNewConstMetric(
					legacyStatusCollectorDescription,
					prometheus.GaugeValue,
					float64(boolToInt(p.ClusterStatus().HasDeletingCondition())),
					m.GetName(),
					providerv1alpha1.StatusClusterTypeDeleting,
				)
			}
		case <-time.After(time.Second):
			return nil
		}
	}
}

func (c *LegacyStatusCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- legacyCreationSecondsCollectorDescription
	ch <- legacyStatusCollectorDescription

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

// timeForCreatedCondition returns the lastTransitionTime for the Created status
// conditions of a legacy provider CR.
//
//     - lastTransitionTime: "2018-11-12T17:01:56.654196952Z"
//       status: "True"
//       type: Created
//
func timeForCreatedCondition(p Provider) time.Time {
	for _, s := range p.ClusterStatus().Conditions {
		if s.Type == "Created" {
			return s.LastTransitionTime.Time
		}
	}

	return time.Time{}
}
