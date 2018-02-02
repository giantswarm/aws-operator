package service

import (
	"testing"

	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"
)

func Test_Service_New(t *testing.T) {
	tests := []struct {
		config               func() Config
		description          string
		expectedErrorHandler func(error) bool
	}{
		{
			description:          "default config is invalid",
			config:               DefaultConfig,
			expectedErrorHandler: IsInvalidConfig,
		},
		{
			description: "production like config is valid",
			config: func() Config {
				config := DefaultConfig()
				config.Logger = microloggertest.New()
				config.Flag = flag.New()
				config.Viper = viper.New()

				config.Description = "test"
				config.GitCommit = "test"
				config.Name = "test"
				config.Source = "test"

				config.Viper.Set(config.Flag.Service.AWS.AccessKey.ID, "accessKeyID")
				config.Viper.Set(config.Flag.Service.AWS.AccessKey.Secret, "accessKeySecret")
				config.Viper.Set(config.Flag.Service.AWS.AccessKey.Session, "session")
				config.Viper.Set(config.Flag.Service.AWS.Region, "myregion")
				config.Viper.Set(config.Flag.Service.AWS.PubKeyFile, "test")

				config.Viper.Set(config.Flag.Service.Installation.Name, "test")

				config.Viper.Set(config.Flag.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				config.Viper.Set(config.Flag.Service.Kubernetes.InCluster, "false")

				return config
			},
			expectedErrorHandler: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := New(tc.config())
			if err != nil {
				if tc.expectedErrorHandler == nil {
					t.Fatalf("unexpected error returned: %v", err)
				}
				if !tc.expectedErrorHandler(err) {
					t.Fatalf("incorrect error returned: %v", err)
				}
			}
		})
	}
}
