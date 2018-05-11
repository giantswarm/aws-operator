package service

import (
	"testing"

	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"
)

func Test_Service_New(t *testing.T) {
	t.Parallel()
	tests := []struct {
		config               func() Config
		description          string
		expectedErrorHandler func(error) bool
	}{
		{
			description:          "default config is invalid",
			config:               func() Config { return Config{} },
			expectedErrorHandler: IsInvalidConfig,
		},
		{
			description: "production like config is valid",
			config: func() Config {
				f := flag.New()

				v := viper.New()
				v.Set(f.Service.AWS.AccessKey.ID, "accessKeyID")
				v.Set(f.Service.AWS.AccessKey.Secret, "accessKeySecret")
				v.Set(f.Service.AWS.AccessKey.Session, "session")
				v.Set(f.Service.AWS.AdvancedMonitoringEC2, true)
				v.Set(f.Service.AWS.S3AccessLogsExpiration, 365)
				v.Set(f.Service.AWS.Region, "myregion")
				v.Set(f.Service.AWS.PubKeyFile, "test")

				v.Set(f.Service.Installation.Name, "test")
				v.Set(f.Service.AWS.LoggingBucket.Delete, true)

				v.Set(f.Service.Kubernetes.Address, "http://127.0.0.1:6443")
				v.Set(f.Service.Kubernetes.InCluster, "false")

				return Config{
					Logger: microloggertest.New(),
					Flag:   f,
					Viper:  v,

					Description: "test",
					GitCommit:   "test",
					ProjectName: "test",
					Source:      "test",
				}
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
