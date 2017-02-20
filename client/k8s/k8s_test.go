package k8s

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestGetRawClientConfig(t *testing.T) {
	tests := []struct {
		name                string
		config              Config
		expectedError       bool
		expectedHost        string
		expectedUsername    string
		expectedPassword    string
		expectedBearerToken string
		expectedCrtFile     string
		expectedKeyFile     string
		expectedCAFile      string
	}{
		{
			name: "Specify only in-cluster config. It should return it. Use basic auth.",
			config: Config{
				InCluster: true,
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:     "http://in-cluster-host",
						Username: "in-cluster-user",
						Password: "in-cluster-password",
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedUsername: "in-cluster-user",
			expectedPassword: "in-cluster-password",
		},
		{
			name: "Specify only in-cluster config. It should return it. Use token auth.",
			config: Config{
				InCluster: true,
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:        "http://in-cluster-host",
						BearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
					}, nil
				},
			},
			expectedHost:        "http://in-cluster-host",
			expectedBearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
		},
		{
			name: "Specify only in-cluster config. It should return it. Use cert auth files.",
			config: Config{
				InCluster: true,
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertFile: "/var/run/kubernetes/client-admin.crt",
							KeyFile:  "/var/run/kubernetes/client-admin.key",
							CAFile:   "/var/run/kubernetes/server-ca.crt",
						},
					}, nil
				},
			},
			expectedHost:    "http://in-cluster-host",
			expectedCrtFile: "/var/run/kubernetes/client-admin.crt",
			expectedKeyFile: "/var/run/kubernetes/client-admin.key",
			expectedCAFile:  "/var/run/kubernetes/server-ca.crt",
		},
		{
			name: "Do not specify anything while using in-cluster config. It should return an error.",
			config: Config{
				InCluster: true,
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedError: true,
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use basic auth.",
			config: Config{
				InCluster: true,
				Host:      "http://host-from-cli",
				Username:  "cli-user",
				Password:  "cli-password",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:     "http://in-cluster-host",
						Username: "in-cluster-user",
						Password: "in-cluster-password",
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedUsername: "in-cluster-user",
			expectedPassword: "in-cluster-password",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use token auth.",
			config: Config{
				InCluster:   true,
				Host:        "http://host-from-cli",
				BearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:        "http://in-cluster-host",
						BearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
					}, nil
				},
			},
			expectedHost:        "http://in-cluster-host",
			expectedBearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use cert auth files.",
			config: Config{
				InCluster: true,
				Host:      "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertFile: "/var/run/kubernetes/client-cli.crt",
					KeyFile:  "/var/run/kubernetes/client-cli.key",
					CAFile:   "/var/run/kubernetes/server-ca-cli.crt",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertFile: "/var/run/kubernetes/client-admin.crt",
							KeyFile:  "/var/run/kubernetes/client-admin.key",
							CAFile:   "/var/run/kubernetes/server-ca.crt",
						},
					}, nil
				},
			},
			expectedHost:    "http://in-cluster-host",
			expectedCrtFile: "/var/run/kubernetes/client-admin.crt",
			expectedKeyFile: "/var/run/kubernetes/client-admin.key",
			expectedCAFile:  "/var/run/kubernetes/server-ca.crt",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the CLI config. Use basic auth.",
			config: Config{
				Host:     "http://host-from-cli",
				Username: "cli-user",
				Password: "cli-password",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:     "http://in-cluster-host",
						Username: "in-cluster-user",
						Password: "in-cluster-password",
					}, nil
				},
			},
			expectedHost:     "http://host-from-cli",
			expectedUsername: "cli-user",
			expectedPassword: "cli-password",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the CLI config. Use token auth.",
			config: Config{
				Host:        "http://host-from-cli",
				BearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:        "http://in-cluster-host",
						BearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
					}, nil
				},
			},
			expectedHost:        "http://host-from-cli",
			expectedBearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the CLI config. Use cert auth files.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertFile: "/var/run/kubernetes/client-cli.crt",
					KeyFile:  "/var/run/kubernetes/client-cli.key",
					CAFile:   "/var/run/kubernetes/server-ca-cli.crt",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertFile: "/var/run/kubernetes/client-admin.crt",
							KeyFile:  "/var/run/kubernetes/client-admin.key",
							CAFile:   "/var/run/kubernetes/server-ca.crt",
						},
					}, nil
				},
			},
			expectedHost:    "http://host-from-cli",
			expectedCrtFile: "/var/run/kubernetes/client-cli.crt",
			expectedKeyFile: "/var/run/kubernetes/client-cli.key",
			expectedCAFile:  "/var/run/kubernetes/server-ca-cli.crt",
		},
		{
			name: "Specify only CLI config. It should return it. Use basic auth.",
			config: Config{
				Host:     "http://host-from-cli",
				Username: "cli-user",
				Password: "cli-password",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:     "http://host-from-cli",
			expectedUsername: "cli-user",
			expectedPassword: "cli-password",
		},
		{
			name: "Specify only CLI config. It should return it. Use token auth.",
			config: Config{
				Host:        "http://host-from-cli",
				BearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:        "http://host-from-cli",
			expectedBearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
		},
		{
			name: "Specify only CLI config. It should return it. Use cert auth files.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertFile: "/var/run/kubernetes/client-cli.crt",
					KeyFile:  "/var/run/kubernetes/client-cli.key",
					CAFile:   "/var/run/kubernetes/server-ca-cli.key",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:    "http://host-from-cli",
			expectedCrtFile: "/var/run/kubernetes/client-cli.crt",
			expectedKeyFile: "/var/run/kubernetes/client-cli.key",
			expectedCAFile:  "/var/run/kubernetes/server-ca-cli.key",
		},
	}
	for _, tc := range tests {
		rawClientConfig, err := getRawClientConfig(tc.config)
		if tc.expectedError {
			assert.Error(t, err, fmt.Sprintf("[%s] An error was expected", tc.name))
			continue
		}
		assert.Nil(t, err, fmt.Sprintf("[%s] An error was unexpected", tc.name))
		assert.Equal(t, tc.expectedHost, rawClientConfig.Host, fmt.Sprintf("[%s] Hosts should be equal", tc.name))
		assert.Equal(t, tc.expectedUsername, rawClientConfig.Username, fmt.Sprintf("[%s] Usernames should be equal", tc.name))
		assert.Equal(t, tc.expectedPassword, rawClientConfig.Password, fmt.Sprintf("[%s] Passwords should be equal", tc.name))
		assert.Equal(t, tc.expectedBearerToken, rawClientConfig.BearerToken, fmt.Sprintf("[%s] Tokens should be equal", tc.name))
		assert.Equal(t, tc.expectedCrtFile, rawClientConfig.TLSClientConfig.CertFile, fmt.Sprintf("[%s] CertFiles should be equal", tc.name))
		assert.Equal(t, tc.expectedKeyFile, rawClientConfig.TLSClientConfig.KeyFile, fmt.Sprintf("[%s] KeyFiles should be equal", tc.name))
		assert.Equal(t, tc.expectedCAFile, rawClientConfig.TLSClientConfig.CAFile, fmt.Sprintf("[%s] CAFiles should be equal", tc.name))
	}
}
