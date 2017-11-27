package cloudformation

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/clustertpr/spec/kubernetes"
)

var (
	defaultCluster = clustertpr.Spec{
		Cluster: spec.Cluster{
			ID: "test-cluster",
		},
		Kubernetes: spec.Kubernetes{
			IngressController: kubernetes.IngressController{
				Domain: "mysubdomain.mydomain.com",
			},
		},
	}
)

func TestAdapterMain(t *testing.T) {
	customObject := awstpr.CustomObject{
		Spec: awstpr.Spec{
			Cluster: defaultCluster,
			AWS: awsspec.AWS{
				Workers: []awsspecaws.Node{
					awsspecaws.Node{},
				},
			},
		},
	}
	clients := Clients{
		EC2: &eC2ClientMock{},
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
			EC2: &eC2ClientMock{},
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
		unexistingSG            bool
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
			unexistingSG:            true,
			expectedError:           true,
			expectedSecurityGroupID: "",
		},
	}

	for _, tc := range testCases {
		a := adapter{}
		clients := Clients{
			EC2: &eC2ClientMock{
				unexistingSg: tc.unexistingSG,
				sgID:         tc.expectedSecurityGroupID,
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
		EC2: &eC2ClientMock{},
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
		description                    string
		customObject                   awstpr.CustomObject
		expectedError                  bool
		expectedAZ                     string
		expectedASGMaxSize             int
		expectedASGMinSize             int
		expectedHealthCheckGracePeriod int
		expectedMaxBatchSize           string
		expectedMinInstancesInService  string
		expectedRollingUpdatePauseTime string
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
					Cluster: defaultCluster,
					AWS: awsspec.AWS{
						AZ: "myaz",
						Workers: []awsspecaws.Node{
							awsspecaws.Node{},
							awsspecaws.Node{},
							awsspecaws.Node{},
						},
					},
				},
			},
			expectedAZ:                     "myaz",
			expectedASGMaxSize:             3,
			expectedASGMinSize:             3,
			expectedHealthCheckGracePeriod: gracePeriodSeconds,
			expectedMaxBatchSize:           strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32),
			expectedMinInstancesInService:  strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32),
			expectedRollingUpdatePauseTime: rollingUpdatePauseTime,
		},
	}

	clients := Clients{
		EC2: &eC2ClientMock{},
	}
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

			if !tc.expectedError {
				if a.AZ != tc.expectedAZ {
					t.Errorf("unexpected output, got %q, want %q", a.AZ, tc.expectedAZ)
				}

				if a.ASGMaxSize != tc.expectedASGMaxSize {
					t.Errorf("unexpected output, got %d, want %d", a.ASGMaxSize, tc.expectedASGMaxSize)
				}

				if a.ASGMinSize != tc.expectedASGMinSize {
					t.Errorf("unexpected output, got %d, want %d", a.ASGMinSize, tc.expectedASGMinSize)
				}

				if a.HealthCheckGracePeriod != tc.expectedHealthCheckGracePeriod {
					t.Errorf("unexpected output, got %d, want %d", a.HealthCheckGracePeriod, tc.expectedHealthCheckGracePeriod)
				}

				if a.MaxBatchSize != tc.expectedMaxBatchSize {
					t.Errorf("unexpected output, got %q, want %q", a.MaxBatchSize, tc.expectedMaxBatchSize)
				}

				if a.MinInstancesInService != tc.expectedMinInstancesInService {
					t.Errorf("unexpected output, got %q, want %q", a.MinInstancesInService, tc.expectedMinInstancesInService)
				}

				if a.RollingUpdatePauseTime != tc.expectedRollingUpdatePauseTime {
					t.Errorf("unexpected output, got %q, want %q", a.RollingUpdatePauseTime, tc.expectedRollingUpdatePauseTime)
				}
			}
		})
	}
}

func TestAdapterAutoScalingGroupLoadBalancerName(t *testing.T) {
	testCases := []struct {
		description              string
		customObject             awstpr.CustomObject
		expectedLoadBalancerName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
				},
			},
			expectedLoadBalancerName: "test-cluster-mysubdomain",
		},
	}

	clients := Clients{
		EC2: &eC2ClientMock{},
	}
	for _, tc := range testCases {
		a := adapter{}
		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.LoadBalancerName != tc.expectedLoadBalancerName {
				t.Errorf("unexpected output, got %q, want %q", a.LoadBalancerName, tc.expectedLoadBalancerName)
			}
		})
	}
}

func TestAdapterAutoScalingGroupSubnetID(t *testing.T) {
	testCases := []struct {
		description                string
		customObject               awstpr.CustomObject
		expectedReceivedSubnetName string
		expectedError              bool
		unexistentSubnet           bool
	}{
		{
			description: "existent subnet",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
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
			expectedReceivedSubnetName: "test-cluster-public",
		},
		{
			description: "unexistent subnet",
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
			unexistentSubnet:           true,
			expectedError:              true,
			expectedReceivedSubnetName: "",
		},
	}

	for _, tc := range testCases {
		a := adapter{}
		clients := Clients{
			EC2: &eC2ClientMock{
				unexistingSubnet: tc.unexistentSubnet,
				clusterID:        "test-cluster",
			},
			IAM: &iAMClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			// the mock does the check internally, the returned subnet id is not related
			// to input
			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}
