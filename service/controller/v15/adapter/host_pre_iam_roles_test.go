package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterHostIAMRolesRegularFields(t *testing.T) {
	t.Parallel()
	guestAccountID := "myGuestAccountID"
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				API: v1alpha1.AWSConfigSpecAWSAPI{
					HostedZones: "apiHostedZones",
				},
				Etcd: v1alpha1.AWSConfigSpecAWSEtcd{
					HostedZones: "etcdHostedZone",
				},
				Ingress: v1alpha1.AWSConfigSpecAWSIngress{
					HostedZones: "ingressHostedZone",
				},
			},
		},
	}

	testCases := []struct {
		description                string
		expectedPeerAccessRoleName string
		expectedGuestAccountID     string
	}{
		{
			description:                "basic matching, all fields present",
			expectedPeerAccessRoleName: "test-cluster-vpc-peer-access",
			expectedGuestAccountID:     guestAccountID,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject:   customObject,
				GuestAccountID: guestAccountID,
			}
			err := a.HostPre.IAMRoles.Adapt(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.HostPre.IAMRoles.PeerAccessRoleName != tc.expectedPeerAccessRoleName {
				t.Errorf("unexpected PeerAccessRoleName, got %q, want %q", a.HostPre.IAMRoles.PeerAccessRoleName, tc.expectedPeerAccessRoleName)
			}
			if a.HostPre.IAMRoles.GuestAccountID != tc.expectedGuestAccountID {
				t.Errorf("unexpected GuestAccountID, got %q, want %q", a.HostPre.IAMRoles.GuestAccountID, tc.expectedGuestAccountID)
			}
		})
	}
}
