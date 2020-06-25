package nodeautorepair

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	gaugeValue float64 = 1

	// prometheusNamespace is the namespace to use for Prometheus metrics.
	// See: https://godoc.org/github.com/prometheus/client_golang/prometheus#Opts
	prometheusNamespace = "giantswarm"

	// prometheusSubsystem is the subsystem to use for Prometheus metrics.
	// See: https://godoc.org/github.com/prometheus/client_golang/prometheus#Opts
	prometheusSubsystem = "node_auto_repair"
)

var (
	nodeAutoRepairTermination = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "aws_operator_node_auto_repair_termination",
			Help:      "Metric representing node termination due node auto repair feature.",
		},
		[]string{"cluster_id", "node_name", "instance_id"},
	)
)

func init() {
	prometheus.MustRegister(nodeAutoRepairTermination)

}

// reportNodeTermination is a utility function for updating metrics related to
// node auto repair node termination.
func reportNodeTermination(clusterID string, nodeName string, instance_id string) {
	nodeAutoRepairTermination.WithLabelValues(
		clusterID, nodeName, instance_id,
	).Set(gaugeValue)
}
