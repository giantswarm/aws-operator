package adapter

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func Test_AdapterLaunchConfiguration_RegularFields(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description                      string
		customObject                     v1alpha1.AWSConfig
		expectedError                    bool
		expectedInstanceType             string
		expectedAssociatePublicIPAddress bool
		expectedBlockDeviceMappings      []BlockDeviceMapping
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								DockerVolumeSizeGB: 250,
								InstanceType:       "myinstancetype",
							},
						},
					},
				},
			},
			expectedInstanceType:             "myinstancetype",
			expectedAssociatePublicIPAddress: false,
			expectedBlockDeviceMappings: []BlockDeviceMapping{
				{
					DeleteOnTermination: true,
					DeviceName:          defaultEBSVolumeMountPoint,
					VolumeSize:          "250",
					VolumeType:          defaultEBSVolumeType,
				},
				{
					DeleteOnTermination: true,
					DeviceName:          logEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
				{
					DeleteOnTermination: true,
					DeviceName:          kubeletEBSVolumeMountPoint,
					VolumeSize:          "250",
					VolumeType:          defaultEBSVolumeType,
				},
			},
		},
		{
			description: "basic matching, verify that default value is used for docker volume when field is missing in CR",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
					AWS: v1alpha1.AWSConfigSpecAWS{
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedInstanceType:             "myinstancetype",
			expectedAssociatePublicIPAddress: false,
			expectedBlockDeviceMappings: []BlockDeviceMapping{
				{
					DeleteOnTermination: true,
					DeviceName:          defaultEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
				{
					DeleteOnTermination: true,
					DeviceName:          logEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
				{
					DeleteOnTermination: true,
					DeviceName:          kubeletEBSVolumeMountPoint,
					VolumeSize:          defaultEBSVolumeSize,
					VolumeType:          defaultEBSVolumeType,
				},
			},
		},
	}
	for _, tc := range testCases {
		a := Adapter{}

		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				StackState: StackState{
					WorkerDockerVolumeSizeGB:  key.WorkerDockerVolumeSizeGB(tc.customObject),
					WorkerKubeletVolumeSizeGB: key.WorkerDockerVolumeSizeGB(tc.customObject),
				},
			}
			err := a.Guest.LaunchConfiguration.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.Guest.LaunchConfiguration.ASGType != key.KindWorker {
				t.Errorf("unexpected ASGType, got %q, want %q", a.Guest.LaunchConfiguration.ASGType, key.KindWorker)
			}
			if a.Guest.LaunchConfiguration.WorkerInstanceType != tc.expectedInstanceType {
				t.Errorf("unexpected InstanceType, got %q, want %q", a.Guest.LaunchConfiguration.WorkerInstanceType, tc.expectedInstanceType)
			}
			if a.Guest.LaunchConfiguration.WorkerAssociatePublicIPAddress != tc.expectedAssociatePublicIPAddress {
				t.Errorf("unexpected WorkerAssociatePublicIPAddress, got %t, want %t", a.Guest.LaunchConfiguration.WorkerAssociatePublicIPAddress, tc.expectedAssociatePublicIPAddress)
			}
			if !reflect.DeepEqual(a.Guest.LaunchConfiguration.WorkerBlockDeviceMappings, tc.expectedBlockDeviceMappings) {
				t.Errorf("unexpected BlockDeviceMappings, got %v, want %v", a.Guest.LaunchConfiguration.WorkerBlockDeviceMappings, tc.expectedBlockDeviceMappings)
			}
		})
	}
}

func Test_AdapterLaunchConfiguration_SmallCloudConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description  string
		expectedLine string
	}{
		{
			description:  "contains S3 URL",
			expectedLine: fmt.Sprintf("s3://000000000000-g8s-test-cluster/ignition/worker"),
		},
	}

	a := Adapter{}
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
			AWS: v1alpha1.AWSConfigSpecAWS{
				Region: "myregion",
				Workers: []v1alpha1.AWSConfigSpecAWSNode{
					{
						ImageID:      "myimageid",
						InstanceType: "myinstancetype",
					},
				},
			},
			VersionBundle: v1alpha1.AWSConfigSpecVersionBundle{
				Version: "0.1.0",
			},
		},
	}
	cfg := Config{
		CustomObject:           customObject,
		TenantClusterAccountID: "000000000000",
	}
	err := a.Guest.LaunchConfiguration.Adapt(cfg)

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	data, err := base64.StdEncoding.DecodeString(a.Guest.LaunchConfiguration.WorkerSmallCloudConfig)
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
