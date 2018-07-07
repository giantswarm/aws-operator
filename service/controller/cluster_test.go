package controller

import (
	"testing"

	versionedfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	apiextensionsclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func newTestClusterConfig() ClusterConfig {
	return ClusterConfig{
		G8sClient:    versionedfake.NewSimpleClientset(),
		K8sClient:    kubernetesfake.NewSimpleClientset(),
		K8sExtClient: apiextensionsclientfake.NewSimpleClientset(),
		Logger:       microloggertest.New(),

		AccessLogsExpiration: 365,
		GuestAWSConfig: ClusterConfigAWSConfig{
			AccessKeyID:     "guest-key",
			AccessKeySecret: "guest-secret",
			Region:          "guest-myregion",
			SessionToken:    "guest-token",
		},
		HostAWSConfig: ClusterConfigAWSConfig{
			AccessKeyID:     "host-key",
			AccessKeySecret: "host-secret",
			Region:          "host-myregion",
			SessionToken:    "host-token",
		},
		InstallationName:    "test",
		DeleteLoggingBucket: true,
		ProjectName:         "aws-operator",
		PubKeyFile:          "~/.ssh/id_rsa.pub",
		SSOPublicKey:        "test",
		EncrypterBackend:    "kms",
	}
}

func Test_NewCluster_Fails_Without_Regions(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description      string
		guestCredentials ClusterConfigAWSConfig
		hostCredentials  ClusterConfigAWSConfig
		expectedError    bool
	}{
		// Test 0.
		{
			description: "misisng region in guest",
			guestCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				SessionToken:    "token",
			},
			hostCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			expectedError: true,
		},
		// Test 1.
		{
			description: "missing region in host",
			guestCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			hostCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				SessionToken:    "token",
			},
			expectedError: true,
		},
		// Test 2.
		{
			description: "region exists in guest and host",
			guestCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			hostCredentials: ClusterConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			expectedError: false,
		},
	}

	for i, tc := range testCases {
		c := newTestClusterConfig()
		c.GuestAWSConfig = tc.guestCredentials
		c.HostAWSConfig = tc.hostCredentials

		_, err := NewCluster(c)
		if tc.expectedError != (err != nil) {
			t.Errorf("case %d: expected error = %v, got = %#v", i, tc.expectedError, err)
		}
	}
}
