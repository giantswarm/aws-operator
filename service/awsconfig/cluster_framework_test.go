package awsconfig

import (
	"testing"

	versionedfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	apiextensionsclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func newTestClusterFrameworkConfig() ClusterFrameworkConfig {
	return ClusterFrameworkConfig{
		G8sClient:    versionedfake.NewSimpleClientset(),
		K8sClient:    kubernetesfake.NewSimpleClientset(),
		K8sExtClient: apiextensionsclientfake.NewSimpleClientset(),
		Logger:       microloggertest.New(),

		GuestAWSConfig: FrameworkConfigAWSConfig{
			AccessKeyID:     "guest-key",
			AccessKeySecret: "guest-secret",
			Region:          "guest-myregion",
			SessionToken:    "guest-token",
		},
		HostAWSConfig: FrameworkConfigAWSConfig{
			AccessKeyID:     "host-key",
			AccessKeySecret: "host-secret",
			Region:          "host-myregion",
			SessionToken:    "host-token",
		},
		InstallationName: "test",
		ProjectName:      "aws-operator",
		PubKeyFile:       "~/.ssh/id_rsa.pub",
	}
}

func Test_NewFramework_Fails_Without_Regions(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description      string
		guestCredentials FrameworkConfigAWSConfig
		hostCredentials  FrameworkConfigAWSConfig
		expectedError    bool
	}{
		// Test 0.
		{
			description: "misisng region in guest",
			guestCredentials: FrameworkConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				SessionToken:    "token",
			},
			hostCredentials: FrameworkConfigAWSConfig{
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
			guestCredentials: FrameworkConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			hostCredentials: FrameworkConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				SessionToken:    "token",
			},
			expectedError: true,
		},
		// Test 2.
		{
			description: "region exists in guest and host",
			guestCredentials: FrameworkConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			hostCredentials: FrameworkConfigAWSConfig{
				AccessKeyID:     "key",
				AccessKeySecret: "secret",
				Region:          "myregion",
				SessionToken:    "token",
			},
			expectedError: false,
		},
	}

	for i, tc := range testCases {
		config := newTestClusterFrameworkConfig()
		config.GuestAWSConfig = tc.guestCredentials
		config.HostAWSConfig = tc.hostCredentials

		_, err := NewClusterFramework(config)
		if tc.expectedError != (err != nil) {
			t.Errorf("case %d: expected error = %v, got = %#v", i, tc.expectedError, err)
		}
	}
}
