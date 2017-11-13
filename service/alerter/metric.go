package alerter

import "github.com/prometheus/client_golang/prometheus"

var (
	orphanResourcesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "awsoperator",
			Subsystem: "resources",
			Name:      "orphan_resources_total",
			Help:      "Number of AWS resources not associated with a cluster to be cleaned up.",
		},
		[]string{"resource"},
	)
)

func init() {
	prometheus.MustRegister(orphanResourcesTotal)
}
