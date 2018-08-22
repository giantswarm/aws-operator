package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/client/aws"
)

const (
	clusterLabel           = "cluster_id"
	statusLabel            = "status"
	listClusterConfigLimit = 20
)

var clustersDesc *prometheus.Desc = prometheus.NewDesc(
	prometheus.BuildFQName(Namespace, "", "cluster_info"),
	"Cluster information.",
	[]string{
		clusterLabel,
		statusLabel,
	},
	nil,
)

func (c *Collector) collectClusterInfo(ch chan<- prometheus.Metric, clients []aws.Clients) {
	c.logger.Log("level", "debug", "message", "collecting metrics for clusters")

	opts := v1.ListOptions{
		Limit: listClusterConfigLimit,
	}

	continueToken := ""

	for {
		opts.Continue = continueToken

		clustersConfigList, err := c.g8sClient.ProviderV1alpha1().AWSConfigs("").List(opts)
		if err != nil {
			c.logger.Log(err)
		}

		for _, clusterConfig := range clustersConfigList.Items {

			clusterID := clusterConfig.Name
			status := "Running"

			if clusterConfig.DeletionTimestamp != nil {
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

		continueToken = clustersConfigList.Continue
		if continueToken == "" {
			break
		}
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for clusters")
}
