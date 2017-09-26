package service

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/giantswarm/kvm-operator/flag"
	"github.com/giantswarm/micrologger/microloggertest"
)

func TestNew(t *testing.T) {
	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config is invalid.
		{
			config:               DefaultConfig,
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test a production-like config is valid.
		{
			config: func() Config {
				config := DefaultConfig()

				config.Logger = microloggertest.New()

				config.Flag = flag.New()
				config.Viper = viper.New()

				config.Description = "test"
				config.GitCommit = "test"
				config.Name = "test"
				config.Source = "test"

				config.Viper.Set(config.Flag.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				config.Viper.Set(config.Flag.Service.Kubernetes.InCluster, "false")

				return config
			},
			expectedErrorHandler: nil,
		},
	}

	for index, test := range tests {
		_, err := New(test.config())

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned: %v", index, err)
			}
			if !test.expectedErrorHandler(err) {
				t.Fatalf("%v: incorrect error returned: %v", index, err)
			}
		}
	}
}
