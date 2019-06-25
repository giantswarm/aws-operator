package key

import (
	"encoding/json"
	"strconv"
	"testing"

	g8sclusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	g8sproviderv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func withg8sMachineDeploymentSpecToCMAMachineDeployment(cr v1alpha1.MachineDeployment, providerExtension g8sclusterv1alpha1.AWSMachineDeploymentSpec) v1alpha1.MachineDeployment {
	var err error

	if cr.Spec.Template.Spec.ProviderSpec.Value == nil {
		cr.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{}
	}

	cr.Spec.Template.Spec.ProviderSpec.Value.Raw, err = json.Marshal(&providerExtension)
	if err != nil {
		panic(err)
	}

	return cr
}

func TestStatusAvailabilityZones(t *testing.T) {
	testCases := []struct {
		name              string
		cr                v1alpha1.MachineDeployment
		providerExtension g8sclusterv1alpha1.AWSMachineDeploymentSpec
		expectedZones     []g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZone
		errorMatcher      func(error) bool
	}{
		{
			name: "case 0: Test with three AZs",
			cr:   v1alpha1.MachineDeployment{},
			expectedZones: []g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
				{
					Name: "eu-central-1c",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var err error // TODO: shall be filled later
			zones := StatusAvailabilityZones(withg8sMachineDeploymentSpecToCMAMachineDeployment(tc.cr, tc.providerExtension))

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
