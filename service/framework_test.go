package service

import (
	"reflect"
	"testing"

	versionedfake "github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/micrologger/microloggertest"
	apiextensionsclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func newTestFrameworkConfig() FrameworkConfig {
	return FrameworkConfig{
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
		Name:             "aws-operator",
		PubKeyFile:       "~/.ssh/id_rsa.pub",
	}
}

func Test_newVersionedResources(t *testing.T) {
	testCases := []struct {
		description       string
		expectedResources map[string][]string
		errorMatcher      func(err error) bool
	}{
		{
			description: "",
			expectedResources: map[string][]string{
				"": []string{
					"legacyv2",
				},
				"0.1.0": []string{
					"legacyv2",
				},
				"0.2.0": []string{
					"kmskey",
					"s3bucketv2",
					"s3objectv2",
					"cloudformationv2",
					"namespacev2",
					"servicev2",
					"endpointsv2",
				},
				"1.0.0": []string{
					"legacyv2",
				},
				"2.0.0": []string{
					"kmskey",
					"s3bucketv2",
					"s3objectv2",
					"cloudformationv2",
					"namespacev2",
					"servicev2",
					"endpointsv2",
				},
				"2.0.1": []string{
					"kmskey",
					"s3bucketv2",
					"s3objectv3",
					"cloudformationv2",
					"namespacev2",
					"servicev2",
					"endpointsv2",
				},
				"2.0.2": []string{
					"kmskey",
					"s3bucketv2",
					"s3objectv3",
					"cloudformationv2",
					"namespacev2",
					"servicev2",
					"endpointsv2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := newTestFrameworkConfig()

			versionedResources, err := newVersionedResources(config)
			if err != nil {
				t.Fatalf("unexpected error %#v", err)
			}

			versionedResourceNames := make(map[string][]string)

			for versionBundleVersion, resources := range versionedResources {
				resourceNames := []string{}

				for _, resource := range resources {
					resourceNames = append(resourceNames, resource.Underlying().Name())
				}

				versionedResourceNames[versionBundleVersion] = resourceNames
			}

			if !reflect.DeepEqual(tc.expectedResources, versionedResourceNames) {
				t.Errorf("expected versioned resources %v got %v", tc.expectedResources, versionedResourceNames)
			}
		})
	}
}

func Test_NewFramework_Fails_Without_Regions(t *testing.T) {
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
		config := newTestFrameworkConfig()
		config.GuestAWSConfig = tc.guestCredentials
		config.HostAWSConfig = tc.hostCredentials

		_, err := NewFramework(config)
		if tc.expectedError != (err != nil) {
			t.Errorf("case %d: expected error = %v, got = %#v", i, tc.expectedError, err)
		}
	}
}
