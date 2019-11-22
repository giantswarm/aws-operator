package version

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

var buildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "giantswarm",
		Name:      "build_info",
		Help:      "A metric with a constant '1' value labeled by commit, golang version, golang os, and golang arch.",
	},
	[]string{"commit", "golang_version", "golang_goos", "golang_goarch", "reconciled_version"},
)

func init() {
	prometheus.MustRegister(buildInfo)
}

func (s *Service) updateBuildInfoMetric() {
	for _, bundle := range s.versionBundles {
		buildInfo.WithLabelValues(s.gitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH, bundle.Version).Set(1)
	}
}
