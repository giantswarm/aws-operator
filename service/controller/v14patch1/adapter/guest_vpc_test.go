package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterVPCRegularFields(t *testing.T) {
	t.Parallel()
	cidr := "10.0.0.0/24"
	peerID := "mypeerID"
	installationName := "myinstallation"
	hostAccountID := "myHostAccountID"

	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				VPC: v1alpha1.AWSConfigSpecAWSVPC{
					CIDR:   cidr,
					PeerID: peerID,
				},
			},
		},
	}
	testCases := []struct {
		description              string
		expectedCIDR             string
		expectedPeerVPCID        string
		expectedInstallationName string
		expectedHostAccountID    string
	}{
		{
			description:              "basic matching, all fields present",
			expectedCIDR:             cidr,
			expectedPeerVPCID:        peerID,
			expectedInstallationName: installationName,
			expectedHostAccountID:    hostAccountID,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject:     customObject,
				Clients:          Clients{},
				InstallationName: installationName,
				HostAccountID:    hostAccountID,
				HostClients: Clients{
					IAM: &IAMClientMock{},
					STS: &STSClientMock{},
				},
			}
			err := a.Guest.VPC.Adapt(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.VPC.CidrBlock != tc.expectedCIDR {
				t.Errorf("unexpected CidrBlock, got %q, want %q", a.Guest.VPC.CidrBlock, tc.expectedCIDR)
			}

			if a.Guest.VPC.PeerVPCID != tc.expectedPeerVPCID {
				t.Errorf("unexpected PeerVPCID, got %q, want %q", a.Guest.VPC.PeerVPCID, tc.expectedPeerVPCID)
			}

			if a.Guest.VPC.InstallationName != tc.expectedInstallationName {
				t.Errorf("unexpected InstallationName, got %q, want %q", a.Guest.VPC.InstallationName, tc.expectedInstallationName)
			}

			if a.Guest.VPC.HostAccountID != tc.expectedHostAccountID {
				t.Errorf("unexpected HostAccountID, got %q, want %q", a.Guest.VPC.HostAccountID, tc.expectedHostAccountID)
			}
		})
	}
}

func TestAdapterVPCPeerRoleField(t *testing.T) {
	t.Parallel()
	peerRoleArn := "myPeerRoleArn"
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}
	testCases := []struct {
		description         string
		expectedPeerRoleArn string
	}{
		{
			description:         "basic matching, all fields present",
			expectedPeerRoleArn: peerRoleArn,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: customObject,
				HostClients: Clients{
					IAM: &IAMClientMock{peerRoleArn: peerRoleArn},
					STS: &STSClientMock{},
				},
			}
			err := a.Guest.VPC.Adapt(cfg)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.VPC.PeerRoleArn != tc.expectedPeerRoleArn {
				t.Errorf("unexpected PeerRoleArn, got %q, want %q", a.Guest.VPC.PeerRoleArn, tc.expectedPeerRoleArn)
			}
		})
	}
}
