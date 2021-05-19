package terminateunhealthynode

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	gaugeValue float64 = 1
)

var (
	nodeUnhealthyNodeTermination = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aws_operator_unhealthy_node_termination",
			Help: "Gauge representing node termination due to terminate unhealthy node feature.",
		},
		[]string{"cluster_id", "terminated_node", "terminated_instance_id"},
	)

	nodeUnhealthyNodeTerminationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aws_operator_unhealthy_node_termination_count",
			Help: "Counter representing node termination count for each cluster.",
		},
		[]string{"cluster_id"},
	)
)

func init() {
	prometheus.MustRegister(nodeUnhealthyNodeTermination)
	prometheus.MustRegister(nodeUnhealthyNodeTerminationCounter)
}

// reportNodeTermination is a utility function for updating metrics related to
// node auto repair node termination.
func reportNodeTermination(clusterID string, nodeName string, instanceID string) {
	nodeUnhealthyNodeTermination.WithLabelValues(
		clusterID, nodeName, instanceID,
	).Set(gaugeValue)

	nodeUnhealthyNodeTerminationCounter.WithLabelValues(clusterID).Inc()
}
