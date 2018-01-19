package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/randomkeytpr"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/fake"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/flag"
	"github.com/giantswarm/aws-operator/service/cloudconfigv1"
	"github.com/giantswarm/aws-operator/service/resource/legacyv1"
	"github.com/giantswarm/aws-operator/service/resource/namespacev1"
)

func Test_Service_NewResourceRouter(t *testing.T) {
	var err error

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

	versionedResources, err := newVersionedResources(config, k8sClient, awsConfig, awsHostConfig)
	if err != nil {
		t.Fatalf("unexpected error %#v", err)
	}

	testCases := []struct {
		description       string
		customObject      v1alpha1.AWSConfig
		expectedResources []string
		errorMatcher      func(err error) bool
		resourceRouter    func(ctx context.Context, obj interface{}) ([]framework.Resource, error)
	}{
		{
			description: "No version in version bundle so return legacy resource",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
						Version: "",
					},
				},
			},
			expectedResources: []string{
				"legacyv2",
			},
			errorMatcher:   nil,
			resourceRouter: NewResourceRouter(versionedResources),
		},
		{
			description: "Legacy version in version bundle so return legacy resource",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
						Version: "1.0.0",
					},
				},
			},
			expectedResources: []string{
				"legacyv2",
			},
			errorMatcher:   nil,
			resourceRouter: NewResourceRouter(versionedResources),
		},
		{
			description: "Invalid version in version bundle returns error",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
						Version: "8.0.8",
					},
				},
			},
			expectedResources: []string{},
			errorMatcher:      IsInvalidVersion,
			resourceRouter:    NewResourceRouter(versionedResources),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resources, err := tc.resourceRouter(context.TODO(), &tc.customObject)
			if err != nil {
				if tc.errorMatcher == nil {
					t.Error("expected", nil, "got", "error matcher")
				} else if !tc.errorMatcher(err) {
					t.Error("expected", true, "got", false)
				}
			} else {
				resourceNames := []string{}

				for _, resource := range resources {
					resourceNames = append(resourceNames, resource.Underlying().Name())
				}

				if !reflect.DeepEqual(tc.expectedResources, resourceNames) {
					t.Errorf("expected %v got %v", tc.expectedResources, resourceNames)
				}
			}
		})
	}
}

func Test_Service_NewTPRResourceRouter(t *testing.T) {
	var err error

	var awsConfig awsclient.Config
	{
		awsConfig = awsclient.Config{
			AccessKeyID:     "accessKeyID",
			AccessKeySecret: "accessKeySecret",
			SessionToken:    "sessionToken",
		}
	}

	var certWatcher *certificatetpr.Service
	{
		certConfig := certificatetpr.DefaultServiceConfig()
		certConfig.K8sClient = fake.NewSimpleClientset()
		certConfig.Logger = microloggertest.New()
		certWatcher, err = certificatetpr.NewService(certConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var keyWatcher *randomkeytpr.Service
	{
		keyConfig := randomkeytpr.DefaultServiceConfig()
		keyConfig.K8sClient = fake.NewSimpleClientset()
		keyConfig.Logger = microloggertest.New()
		keyWatcher, err = randomkeytpr.NewService(keyConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var ccService *cloudconfigv1.CloudConfig
	{
		ccServiceConfig := cloudconfigv1.DefaultConfig()
		ccServiceConfig.Logger = microloggertest.New()
		ccServiceConfig.Logger = microloggertest.New()

		ccService, err = cloudconfigv1.New(ccServiceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	var legacyResource framework.Resource
	{
		legacyConfig := legacyv1.DefaultConfig()
		legacyConfig.AwsConfig = awsConfig
		legacyConfig.AwsHostConfig = awsConfig
		legacyConfig.CertWatcher = certWatcher
		legacyConfig.CloudConfig = ccService
		legacyConfig.InstallationName = "test"
		legacyConfig.K8sClient = fake.NewSimpleClientset()
		legacyConfig.KeyWatcher = keyWatcher
		legacyConfig.Logger = microloggertest.New()
		legacyConfig.PubKeyFile = "test"

		legacyResource, err = legacyv1.New(legacyConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	namespaceConfig := namespacev1.DefaultConfig()
	namespaceConfig.K8sClient = fake.NewSimpleClientset()
	namespaceConfig.Logger = microloggertest.New()

	namespaceResource, err := namespacev1.New(namespaceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	allResources := []framework.Resource{
		legacyResource,
		namespaceResource,
	}
	legacyResources := []framework.Resource{
		legacyResource,
	}

	versionedResources := make(map[string][]framework.Resource)
	versionedResources["0.1.0"] = legacyResources
	versionedResources["0.2.0"] = allResources
	versionedResources["1.0.0"] = legacyResources

	testCases := []struct {
		customObject      awstpr.CustomObject
		expectedResources []framework.Resource
		errorMatcher      func(err error) bool
		resourceRouter    func(ctx context.Context, obj interface{}) ([]framework.Resource, error)
	}{
		// Case 0. No version in version bundle so return legacy resources.
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					VersionBundle: spec.VersionBundle{
						Version: "",
					},
				},
			},
			expectedResources: legacyResources,
			errorMatcher:      nil,
			resourceRouter:    NewTPRResourceRouter(versionedResources),
		},
		// Case 1. Legacy version in version bundle so return legacy resources.
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					VersionBundle: spec.VersionBundle{
						Version: "0.1.0",
					},
				},
			},
			expectedResources: legacyResources,
			errorMatcher:      nil,
			resourceRouter:    NewTPRResourceRouter(versionedResources),
		},
		// Case 2. Cloud formation version in version bundle so return all resources.
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					VersionBundle: spec.VersionBundle{
						Version: "0.2.0",
					},
				},
			},
			expectedResources: allResources,
			errorMatcher:      nil,
			resourceRouter:    NewTPRResourceRouter(versionedResources),
		},
		// Case 3. Kubernetes update to 1.8.4.
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					VersionBundle: spec.VersionBundle{
						Version: "1.0.0",
					},
				},
			},
			expectedResources: legacyResources,
			errorMatcher:      nil,
			resourceRouter:    NewTPRResourceRouter(versionedResources),
		},
		// Case 4. Invalid version returns an error.
		{
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					VersionBundle: spec.VersionBundle{
						Version: "0.3.0",
					},
				},
			},
			expectedResources: allResources,
			errorMatcher:      IsInvalidVersion,
			resourceRouter:    NewTPRResourceRouter(versionedResources),
		},
	}

	for i, tc := range testCases {
		resources, err := tc.resourceRouter(context.TODO(), &tc.customObject)
		if err != nil {
			if tc.errorMatcher == nil {
				t.Fatal("test", i, "expected", nil, "got", "error matcher")
			} else if !tc.errorMatcher(err) {
				t.Fatal("test", i, "expected", true, "got", false)
			}
		} else {
			if !reflect.DeepEqual(tc.expectedResources, resources) {
				t.Fatal("test", i, "expected", tc.expectedResources, "got", resources)
			}
		}
	}
}
