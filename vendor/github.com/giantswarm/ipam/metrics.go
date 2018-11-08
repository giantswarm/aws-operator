package ipam

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "ipam"
)

var (
	subnetCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Name:      "subnet_total",
			Help:      "Number of total subnets.",
		},
	)

	subnetOperationDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Name:      "subnet_operation_duration_milliseconds",
			Help:      "Time taken for subnet operations, in milliseconds.",
		},
		[]string{"operation_name"},
	)
	subnetOperationTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Name:      "subnet_operation_total",
			Help:      "Total number of subnet operations.",
		},
		[]string{"operation_name"},
	)
)

func init() {
	prometheus.MustRegister(subnetCounter)

	prometheus.MustRegister(subnetOperationDuration)
	prometheus.MustRegister(subnetOperationTotal)
}

func updateMetrics(name string, startTime time.Time) {
	subnetOperationDuration.WithLabelValues(name).Set(
		float64(time.Since(startTime) / time.Millisecond),
	)
	subnetOperationTotal.WithLabelValues(name).Inc()
}
