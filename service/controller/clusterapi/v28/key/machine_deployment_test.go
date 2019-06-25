package key

import (
	"encoding/json"
	"strconv"
	"testing"

	g8sclusterv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	g8sproviderv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			name: "case 0: Test with two AZs",
			cr: v1alpha1.MachineDeployment{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						annotation.MachineDeploymentSubnet: "10.100.4.0/24",
					},
				},
			},
			providerExtension: g8sclusterv1alpha1.AWSMachineDeploymentSpec{
				Provider: g8sclusterv1alpha1.AWSMachineDeploymentSpecProvider{
					AvailabilityZones: []string{
						"eu-central-1a",
						"eu-central-1b",
					},
				},
			},
			expectedZones: []g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.0/26",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.64/26",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.128/26",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.192/26",
						},
					},
				},
			},
		},
		{
			name: "case 1: Test with three AZs",
			cr: v1alpha1.MachineDeployment{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						annotation.MachineDeploymentSubnet: "10.100.4.0/24",
					},
				},
			},
			providerExtension: g8sclusterv1alpha1.AWSMachineDeploymentSpec{
				Provider: g8sclusterv1alpha1.AWSMachineDeploymentSpecProvider{
					AvailabilityZones: []string{
						"eu-central-1a",
						"eu-central-1b",
						"eu-central-1c",
					},
				},
			},
			expectedZones: []g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZone{
				{
					Name: "eu-central-1a",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.0/27",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.32/27",
						},
					},
				},
				{
					Name: "eu-central-1b",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.64/27",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.96/27",
						},
					},
				},
				{
					Name: "eu-central-1c",
					Subnet: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnet{
						Private: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPrivate{
							CIDR: "10.100.4.128/27",
						},
						Public: g8sproviderv1alpha1.AWSConfigStatusAWSAvailabilityZoneSubnetPublic{
							CIDR: "10.100.4.160/27",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			zones, err := StatusAvailabilityZones(withg8sMachineDeploymentSpecToCMAMachineDeployment(tc.cr, tc.providerExtension))

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
