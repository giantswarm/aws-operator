package alerter

import "github.com/prometheus/client_golang/prometheus"

var (
	duplicateResourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "awsoperator",
			Subsystem: "resources",
			Name:      "duplicate_resources_total",
			Help:      "Number of clusters with duplicate resources to be cleaned up.",
		},
		[]string{"resource"},
	)
	orphanResourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "awsoperator",
			Subsystem: "resources",
			Name:      "orphan_resources_total",
			Help:      "Number of AWS resources not associated with a cluster to be cleaned up.",
		},
		[]string{"resource"},
	)
	orphanClustersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "awsoperator",
			Subsystem: "resources",
			Name:      "orphan_clusters_total",
			Help:      "Number of clusters without AWS resources associated.",
		},
		[]string{"resource"},
	)
)

func init() {
	prometheus.MustRegister(duplicateResourcesTotal)
	prometheus.MustRegister(orphanResourcesTotal)
	prometheus.MustRegister(orphanClustersTotal)
}
