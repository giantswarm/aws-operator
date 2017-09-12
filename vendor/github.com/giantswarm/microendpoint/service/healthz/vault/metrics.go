package vault

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "microendpoint"
	prometheusSubsystem = "vault"
)

var (
	vaultPermissionDenied = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "permission_denied",
		Help:      "Binary gauge for vault permission denied error.",
	})

	vaultUnknownError = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "unknown_error",
		Help:      "Binary gauge for vault unknown error.",
	})
)

func init() {
	prometheus.MustRegister(vaultPermissionDenied)
	prometheus.MustRegister(vaultUnknownError)
}

func setVaultPermissionDenied() {
	vaultPermissionDenied.Set(1)
}

func setVaultUnknownError() {
	vaultUnknownError.Set(1)
}

func setVaultOK() {
	vaultPermissionDenied.Set(0)
	vaultUnknownError.Set(0)
}
