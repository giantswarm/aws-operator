package healthz

import "github.com/prometheus/client_golang/prometheus"

const (
	// PrometheusFailedLabel allows filtering for failed healthchecks.
	PrometheusFailedLabel string = "failed"
	// PrometheusSuccessfulLabel allows filtering for successful healthchecks.
	PrometheusSuccessfulLabel string = "successful"
)

var (
	healthCheckRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "health_check_request_total",
			Help: "Number of health check requests.",
		},
		[]string{"success"},
	)
	healthCheckRequestTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "health_check_request_milliseconds",
		Help: "Time taken to respond to health check, in milliseconds.",
	})
)

func init() {
	prometheus.MustRegister(healthCheckRequests)
	prometheus.MustRegister(healthCheckRequestTime)
}
