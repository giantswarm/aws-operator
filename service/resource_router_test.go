package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/randomkeytpr"
	"k8s.io/client-go/kubernetes/fake"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/cloudconfigv1"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv1"
	"github.com/giantswarm/aws-operator/service/resource/legacyv1"
	"github.com/giantswarm/aws-operator/service/resource/namespacev1"
)

func Test_Service_NewResourceRouter(t *testing.T) {
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

	var cloudformationResource framework.Resource
	{
		cloudformationConfig := cloudformationv1.DefaultConfig()
		cloudformationConfig.Logger = microloggertest.New()

		cloudformationResource, err = cloudformationv1.New(cloudformationConfig)
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
		cloudformationResource,
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
			resourceRouter:    NewResourceRouter(versionedResources),
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
			resourceRouter:    NewResourceRouter(versionedResources),
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
			resourceRouter:    NewResourceRouter(versionedResources),
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
			resourceRouter:    NewResourceRouter(versionedResources),
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
			resourceRouter:    NewResourceRouter(versionedResources),
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
