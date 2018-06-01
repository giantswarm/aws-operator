package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterLabel = "cluster_id"
	statusLabel  = "status"
)

var (
	clustersDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cluster_info"),
		"Cluster information.",
		[]string{
			clusterLabel,
			statusLabel,
		},
		nil,
	)
)

func (c *Collector) collectClusterInfo(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics for clusters")

	opts := v1.ListOptions{}
	clusters, _ := c.g8sClient.ProviderV1alpha1().AWSConfigs("").List(opts)

	for _, cluster := range clusters.Items {

		clusterID := cluster.Name
		status := "Running"

		_, ok := cluster.Annotations["deletionTimestamp"]
		if ok {
			status = "Terminating"
		}

		ch <- prometheus.MustNewConstMetric(
			clustersDesc,
			prometheus.GaugeValue,
			GaugeValue,
			clusterID,
			status,
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for clusters")
}
