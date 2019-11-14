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
	[]string{"bundle_version", "commit", "golang_version", "golang_goos", "golang_goarch"},
)

func init() {
	prometheus.MustRegister(buildInfo)
}

func (s *Service) updateBuildInfoMetric() {
	buildInfo.WithLabelValues(s.versionBundles[0].Version, s.gitCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH).Set(1)
}
