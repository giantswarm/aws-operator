package service

import (
	"testing"

	cloudformationresource "github.com/giantswarm/aws-operator/service/resource/cloudformation"
	namespaceresource "github.com/giantswarm/aws-operator/service/resource/namespace"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Service_filterResourcesByName(t *testing.T) {
	cloudformationConfig := cloudformationresource.DefaultConfig()
	cloudformationConfig.Logger = microloggertest.New()

	cloudformationResource, err := cloudformationresource.New(cloudformationConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	namespaceConfig := namespaceresource.DefaultConfig()
	namespaceConfig.K8sClient = fake.NewSimpleClientset()
	namespaceConfig.Logger = microloggertest.New()

	namespaceResource, err := namespaceresource.New(namespaceConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	allResources := []framework.Resource{
		cloudformationResource,
		namespaceResource,
	}

	metricsWrapConfig := metricsresource.DefaultWrapConfig()
	metricsWrapConfig.Name = "wrapped"
	wrappedResources, err := metricsresource.Wrap(allResources, metricsWrapConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	retryWrapConfig := retryresource.DefaultWrapConfig()
	wrappedResources, err = retryresource.Wrap(wrappedResources, retryWrapConfig)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	testCases := []struct {
		resources         []framework.Resource
		resourceNames     []string
		expectedResources []framework.Resource
	}{
		// Case 0. Filter by single resource name.
		{
			resources: allResources,
			resourceNames: []string{
				"namespace",
			},
			expectedResources: []framework.Resource{
				namespaceResource,
			},
		},
		// Case 1. Filter by multiple resource names.
		{
			resources: allResources,
			resourceNames: []string{
				"cloudformation",
				"namespace",
			},
			expectedResources: []framework.Resource{
				cloudformationResource,
				namespaceResource,
			},
		},
		// Case 2. Filter by missing resource name.
		{
			resources: allResources,
			resourceNames: []string{
				"legacy",
			},
			expectedResources: []framework.Resource{},
		},
		// Case 3. Filter wrapped resources by resource name.
		{
			resources: wrappedResources,
			resourceNames: []string{
				"namespace",
			},
			expectedResources: []framework.Resource{
				namespaceResource,
			},
		},
		// Case 4. Filter wrapped resources by multiple resource names.
		{
			resources: wrappedResources,
			resourceNames: []string{
				"cloudformation",
				"namespace",
			},
			expectedResources: []framework.Resource{
				cloudformationResource,
				namespaceResource,
			},
		},
		// Case 5. Filter wrapped resources by missing resource name.
		{
			resources: wrappedResources,
			resourceNames: []string{
				"legacy",
			},
			expectedResources: []framework.Resource{},
		},
	}

	for i, tc := range testCases {
		result := filterResourcesByName(tc.resources, tc.resourceNames...)
		if len(result) != len(tc.expectedResources) {
			t.Fatal("test", i, "expected", len(tc.expectedResources), "resources got", len(result))
		}

		resourceNames := make(map[string]bool)

		for _, resource := range result {
			resourceNames[resource.Name()] = true
			resourceNames[resource.Underlying().Name()] = true
		}

		for _, resource := range tc.expectedResources {
			resourceName := resource.Name()
			underlyingName := resource.Underlying().Name()

			if _, ok := resourceNames[resourceName]; !ok {
				t.Fatal("test", i, "expected resource", resourceName, "not found")
			}
			if _, ok := resourceNames[underlyingName]; !ok {
				t.Fatal("test", i, "expected underlying resource", underlyingName, "not found")
			}
		}
	}
}
