package key

import (
	"strconv"
	"testing"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func TestStatusAvailabilityZones(t *testing.T) {
	testCases := []struct {
		name          string
		cr            v1alpha1.MachineDeployment
		expectedZones []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: Test with three AZs",
			cr:   nil,
			expectedZones: []g8sv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
				{
					Name: "eu-central-1c",
					Subnet: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cr := v1alpha1.MachineDeployment{
				Spec: v1alpha1.MachineDeploymentSpec{
					Template: v1alpha1.MachineTemplateSpec{
						Spec: v1alpha1.MachineSpec{
							ProviderSpec: v1alpha1.ProviderSpec{
								Value: &runtime.RawExtension{
									Raw: nil,
								},
							},
						},
					},
				},
			}
			var err error // TODO: shall be filled later
			zones := StatusAvailabilityZones(tc.cr)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			diff := cmp.Diff(zones, tc.expectedZones)
			if diff != "" {
				t.Fatalf("\n\n%s\n", diff)
			}
		})
	}
}
