package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/operatorkit/framework"
)

func Test_Service_NewResourceRouter(t *testing.T) {
	versionedResources, err := newVersionedResources(newTestFrameworkConfig())
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
