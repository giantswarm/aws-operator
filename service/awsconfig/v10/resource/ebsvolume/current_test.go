package ebsvolume

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
)

func Test_CurrentState(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		obj           *v1alpha1.AWSConfig
		expectedState *EBSVolumeState
		ebsVolumes    []ebs.EBSVolumeMock
	}{
		{
			description: "case 0: basic match with no ebs volumes",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{},
			},
		},
		{
			description: "case 1: basic match with ebs volume",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						Attachments: []ebs.VolumeAttachment{},
						VolumeID:    "vol-1234",
					},
				},
			},
			ebsVolumes: []ebs.EBSVolumeMock{
				{
					VolumeID: "vol-1234",
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "case 2: basic match with multiple ebs volumes",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						Attachments: []ebs.VolumeAttachment{},
						VolumeID:    "vol-1234",
					},
					{
						Attachments: []ebs.VolumeAttachment{},
						VolumeID:    "vol-5678",
					},
				},
			},
			ebsVolumes: []ebs.EBSVolumeMock{
				{
					VolumeID: "vol-1234",
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					VolumeID: "vol-5678",
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-5678"),
						},
					},
				},
			},
		},
		{
			description: "case 3: no match due to cluster tag",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{},
			},
			ebsVolumes: []ebs.EBSVolumeMock{
				{
					VolumeID: "vol-1234",
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/other-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
			},
		},
		{
			description: "case 4: no match due to missing pvc tag",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{},
			},
			ebsVolumes: []ebs.EBSVolumeMock{
				{
					VolumeID: "vol-1234",
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
					},
				},
			},
		},
		{
			description: "case 5: multiple ebs volumes with attachments",
			obj:         customObject,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						Attachments: []ebs.VolumeAttachment{
							{
								InstanceID: "i-12345",
								Device:     "/dev/sdh",
							},
						},
						VolumeID: "vol-1234",
					},
					{
						Attachments: []ebs.VolumeAttachment{
							{
								InstanceID: "i-56789",
								Device:     "/dev/sdh",
							},
						},
						VolumeID: "vol-5678",
					},
				},
			},
			ebsVolumes: []ebs.EBSVolumeMock{
				{
					VolumeID: "vol-1234",
					Attachments: []*ec2.VolumeAttachment{
						{
							Device:     aws.String("/dev/sdh"),
							InstanceId: aws.String("i-12345"),
						},
					},
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-1234"),
						},
					},
				},
				{
					VolumeID: "vol-5678",
					Attachments: []*ec2.VolumeAttachment{
						{
							Device:     aws.String("/dev/sdh"),
							InstanceId: aws.String("i-56789"),
						},
					},
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/created-for/pv/name"),
							Value: aws.String("pvc-5678"),
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ebsConfig := ebs.Config{
				Client: &ebs.EC2ClientMock{
					CustomObject: *tc.obj,
					EBSVolumes:   tc.ebsVolumes,
				},
				Logger: microloggertest.New(),
			}
			ebsService, err := ebs.New(ebsConfig)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			c := Config{
				Logger:  microloggertest.New(),
				Service: ebsService,
			}
			newResource, err := New(c)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(*EBSVolumeState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if !reflect.DeepEqual(currentState, tc.expectedState) {
				t.Errorf("expected current state '%#v', got '%#v'", tc.expectedState, currentState)
			}
		})
	}
}
