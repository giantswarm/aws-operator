package s3bucket

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "aws_operator"
	prometehusSubsystem = "resource_s3bucketv22"
)

var (
	s3ObjectsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometehusSubsystem,
			Name:      "s3_objects_total",
			Help:      "Total number of objects within a S3 bucket.",
		},
		[]string{"cluster_id", "bucket_name"},
	)
)

func init() {
	prometheus.MustRegister(s3ObjectsTotal)
}
