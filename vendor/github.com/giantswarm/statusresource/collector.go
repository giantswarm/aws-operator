package statusresource

import (
	"context"
	"fmt"
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
	clusterStatusDescription *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName("statusresource", "cluster", "status"),
		"Cluster status condition as provided by the CR status.",
		[]string{
			"cluster_id",
			"status",
		},
		nil,
	)
)

type CollectorConfig struct {
	Logger  micrologger.Logger
	Watcher func(opts metav1.ListOptions) (watch.Interface, error)
}

type Collector struct {
	logger  micrologger.Logger
	watcher func(opts metav1.ListOptions) (watch.Interface, error)

	bootOnce sync.Once
}

func NewCollector(config CollectorConfig) (*Collector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Watcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Watcher must not be empty", config)
	}

	c := &Collector{
		logger:  config.Logger,
		watcher: config.Watcher,

		bootOnce: sync.Once{},
	}

	return c, nil
}

func (c *Collector) Boot(ctx context.Context) {
	c.bootOnce.Do(func() {
		c.logger.LogCtx(ctx, "level", "debug", "message", "registering collector")

		err := prometheus.Register(prometheus.Collector(c))
		if IsAlreadyRegisteredError(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", "collector already registered")
		} else if err != nil {
			c.logger.Log("level", "error", "message", "registering collector failed", "stack", fmt.Sprintf("%#v", err))
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", "registered collector")
		}
	})
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "start collecting metrics")

	watcher, err := c.watcher(metav1.ListOptions{})
	if err != nil {
		c.logger.Log("level", "error", "message", "watching CRs failed", "stack", fmt.Sprintf("%#v", err))
		return
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
				c.logger.Log("level", "error", "message", "getting meta accessor failed", "stack", fmt.Sprintf("%#v", err))
				break
			}
			p, ok := event.Object.(Provider)
			if !ok {
				c.logger.Log("level", "error", "message", "asserting Provider interface failed")
				break
			}

			ch <- prometheus.MustNewConstMetric(
				clusterStatusDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasCreatingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeCreating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasCreatedCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeCreated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasUpdatingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeUpdating,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasUpdatedCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeUpdated,
			)
			ch <- prometheus.MustNewConstMetric(
				clusterStatusDescription,
				prometheus.GaugeValue,
				float64(boolToInt(p.ClusterStatus().HasDeletingCondition())),
				m.GetName(),
				providerv1alpha1.StatusClusterTypeDeleting,
			)
		case <-time.After(time.Second):
			c.logger.Log("level", "debug", "message", "finished collecting metrics")
			return
		}
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clusterStatusDescription
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
