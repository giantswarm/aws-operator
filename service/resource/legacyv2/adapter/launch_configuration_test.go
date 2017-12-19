package adapter

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterLaunchConfigurationRegularFields(t *testing.T) {
	testCases := []struct {
		description                      string
		customObject                     v1alpha1.AWSConfig
		expectedError                    bool
		expectedImageID                  string
		expectedInstanceType             string
		expectedIAMInstanceProfileName   string
		expectedAssociatePublicIPAddress bool
		expectedBlockDeviceMappings      []BlockDeviceMapping
	}{
		{
			description:   "empty custom object",
			customObject:  v1alpha1.AWSConfig{},
			expectedError: true,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedImageID:                  "myimageid",
			expectedInstanceType:             "myinstancetype",
			expectedIAMInstanceProfileName:   "test-cluster-worker-EC2-K8S-Role",
			expectedAssociatePublicIPAddress: false,
			expectedBlockDeviceMappings: []BlockDeviceMapping{
				BlockDeviceMapping{
					DeleteOnTermination: true,
					DeviceName:          defaultEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
			},
		},
	}
	for _, tc := range testCases {
		clients := Clients{
			EC2: &EC2ClientMock{},
			IAM: &IAMClientMock{},
		}
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getLaunchConfiguration(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.ImageID != tc.expectedImageID {
				t.Errorf("unexpected ImageID, got %q, want %q", a.ImageID, tc.expectedImageID)
			}
			if a.InstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected InstanceType, got %q, want %q", a.InstanceType, tc.expectedInstanceType)
			}
			if a.IAMInstanceProfileName != tc.expectedIAMInstanceProfileName {
				t.Errorf("unexpected IAMInstanceProfileName, got %q, want %q", a.IAMInstanceProfileName, tc.expectedIAMInstanceProfileName)
			}
			if a.AssociatePublicIPAddress != tc.expectedAssociatePublicIPAddress {
				t.Errorf("unexpected IAMInstanceProfileName, got %q, want %q", a.IAMInstanceProfileName, tc.expectedIAMInstanceProfileName)
			}
			if !reflect.DeepEqual(a.BlockDeviceMappings, tc.expectedBlockDeviceMappings) {
				t.Errorf("unexpected BlockDeviceMappings, got %v, want %v", a.BlockDeviceMappings, tc.expectedBlockDeviceMappings)
			}
		})
	}
}

func TestAdapterLaunchConfigurationSecurityGroupID(t *testing.T) {
	testCases := []struct {
		description             string
		customObject            v1alpha1.AWSConfig
		expectedSecurityGroupID string
		expectedError           bool
		unexistingSG            bool
	}{
		{
			description: "existent security group",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedSecurityGroupID: "test-cluster-worker",
		},
		{
			description: "unexistent security group",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			unexistingSG:            true,
			expectedError:           true,
			expectedSecurityGroupID: "",
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{
			EC2: &EC2ClientMock{
				unexistingSg: tc.unexistingSG,
				sgID:         tc.expectedSecurityGroupID,
			},
			IAM: &IAMClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getLaunchConfiguration(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.SecurityGroupID != tc.expectedSecurityGroupID {
				t.Errorf("unexpected SecurityGroupID, got %q, want %q", a.SecurityGroupID, tc.expectedSecurityGroupID)
			}
		})
	}
}

func TestAdapterLaunchConfigurationSmallCloudConfig(t *testing.T) {
	testCases := []struct {
		description  string
		expectedLine string
	}{
		{
			description:  "userdata file",
			expectedLine: "USERDATA_FILE=worker",
		},
		{
			description:  "s3 http uri",
			expectedLine: `s3_http_uri="https://s3.myregion.amazonaws.com/000000000000-g8s-test-cluster/cloudconfig/myversion/$USERDATA_FILE"`,
		},
	}

	a := Adapter{}
	clients := Clients{
		EC2: &EC2ClientMock{},
		IAM: &IAMClientMock{accountID: "000000000000"},
	}
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID:      "test-cluster",
				Version: "myversion",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "myregion",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					v1alpha1.AWSConfigSpecAWSNode{
						ImageID:      "myimageid",
						InstanceType: "myinstancetype",
					},
				},
			},
		},
	}

	err := a.getLaunchConfiguration(customObject, clients)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	data, err := base64.StdEncoding.DecodeString(a.SmallCloudConfig)
	if err != nil {
		t.Errorf("unexpected error decoding SmallCloudConfig %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !strings.Contains(string(data), tc.expectedLine) {
				t.Errorf("SmallCloudConfig didn't contain expected %q, complete: %q", tc.expectedLine, string(data))
			}
		})
	}
}
