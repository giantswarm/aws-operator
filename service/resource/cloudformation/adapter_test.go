package cloudformation

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

func TestAdapterMain(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			AWS: awsspec.AWS{
				Workers: []awsspecaws.Node{
					awsspecaws.Node{},
				},
			},
		},
	}
	clients := Clients{
		EC2: &eC2ClientMock{sgExists: true},
		IAM: &iAMClientMock{},
	}

	a, err := newAdapter(customObject, clients)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expected := prefixWorker
	actual := a.ASGType

	if expected != actual {
		t.Errorf("unexpected value, expecting %q, got %q", expected, actual)
	}
}

func TestAdapterLaunchConfigurationRegularFields(t *testing.T) {
	testCases := []struct {
		description                      string
		customObject                     awstpr.CustomObject
		expectedError                    bool
		expectedImageID                  string
		expectedInstanceType             string
		expectedIAMInstanceProfileName   string
		expectedAssociatePublicIPAddress bool
		expectedBlockDeviceMappings      []BlockDeviceMapping
	}{
		{
			description:   "empty custom object",
			customObject:  awstpr.CustomObject{},
			expectedError: true,
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "test-cluster",
						},
					},
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
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
			expectedAssociatePublicIPAddress: true,
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
			EC2: &eC2ClientMock{sgExists: true},
			IAM: &iAMClientMock{},
		}
		a := adapter{}

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
			if !reflect.DeepEqual(a.BlockDeviceMappings, tc.expectedBlockDeviceMappings) {
				t.Errorf("unexpected BlockDeviceMappings, got %v, want %v", a.BlockDeviceMappings, tc.expectedBlockDeviceMappings)
			}
		})
	}
}

func TestAdapterLaunchConfigurationSecurityGroupID(t *testing.T) {
	testCases := []struct {
		description             string
		customObject            awstpr.CustomObject
		expectedSecurityGroupID string
		expectedError           bool
		securityGroupExists     bool
	}{
		{
			description: "existent security group",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "test-cluster",
						},
					},
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			securityGroupExists:     true,
			expectedSecurityGroupID: "test-cluster-worker",
		},
		{
			description: "unexistent security group",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "test-cluster",
						},
					},
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			securityGroupExists:     false,
			expectedError:           true,
			expectedSecurityGroupID: "",
		},
	}

	for _, tc := range testCases {
		a := adapter{}
		clients := Clients{
			EC2: &eC2ClientMock{
				sgExists: tc.securityGroupExists,
				sgID:     tc.expectedSecurityGroupID,
			},
			IAM: &iAMClientMock{},
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
			expectedLine: `s3_http_uri="https://s3.myregion.amazonaws.com/000000000000-g8s-test-cluster/cloudconfig/$USERDATA_FILE"`,
		},
	}

	a := adapter{}
	clients := Clients{
		EC2: &eC2ClientMock{sgExists: true},
		IAM: &iAMClientMock{accountID: "000000000000"},
	}
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: clustertpr.Spec{
				Cluster: spec.Cluster{
					ID: "test-cluster",
				},
			},
			AWS: awsspec.AWS{
				Region: "myregion",
				Workers: []awsspecaws.Node{
					awsspecaws.Node{
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

func TestAdapterAutoScalingGroupRegularFields(t *testing.T) {
	testCases := []struct {
		description   string
		customObject  awstpr.CustomObject
		expectedError bool
		expectedAZ    string
	}{
		{
			description:  "empty custom object",
			customObject: awstpr.CustomObject{},
			expectedAZ:   "",
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					AWS: awsspec.AWS{
						AZ: "myaz",
					},
				},
			},
			expectedAZ: "myaz",
		},
	}

	clients := Clients{}
	for _, tc := range testCases {
		a := adapter{}
		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.AZ != tc.expectedAZ {
				t.Errorf("unexpected output, got %q, want %q", a.AZ, tc.expectedAZ)
			}
		})
	}
}

func TestValidAmazonAccountID(t *testing.T) {
	tests := []struct {
		name            string
		amazonAccountID string
		err             error
	}{
		{
			name:            "ID has wrong length",
			amazonAccountID: "foo",
			err:             wrongAmazonAccountIDLengthError,
		},
		{
			name:            "ID contains letters",
			amazonAccountID: "123foo123foo",
			err:             malformedAmazonAccountIDError,
		},
		{
			name:            "ID is empty",
			amazonAccountID: "",
			err:             emptyAmazonAccountIDError,
		},
		{
			name:            "ID has correct format",
			amazonAccountID: "123456789012",
			err:             nil,
		},
	}

	for _, tc := range tests {
		err := ValidateAccountID(tc.amazonAccountID)
		assert.Equal(t, microerror.Cause(tc.err), microerror.Cause(err), fmt.Sprintf("[%s] The return value was not what we expected", tc.name))
	}
}
