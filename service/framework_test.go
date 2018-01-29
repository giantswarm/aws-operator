package service

import (
	"reflect"
	"testing"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Service_NewVersionedResources(t *testing.T) {
	config := DefaultConfig()
	config.Name = "aws-operator"
	config.Logger = microloggertest.New()
	config.Flag = flag.New()
	config.Viper = viper.New()
	config.Viper.Set(config.Flag.Service.Installation.Name, "test")
	config.Viper.Set(config.Flag.Service.AWS.PubKeyFile, "~/.ssh/id_rsa.pub")

	k8sClient := fake.NewSimpleClientset()
	awsConfig := awsclient.Config{
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
	}
	awsHostConfig := awsclient.Config{
		AccessKeyID:     "key",
		AccessKeySecret: "secret",
	}

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
					"kmskeyv2",
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
					"kmskeyv2",
					"s3bucketv2",
					"s3objectv2",
					"cloudformationv2",
					"namespacev2",
					"servicev2",
					"endpointsv2",
				},
				"2.1.0": []string{
					"kmskeyv2",
					"s3bucketv2",
					"s3objectv2",
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
			versionedResources, err := NewVersionedResources(config, k8sClient, awsConfig, awsHostConfig)
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
