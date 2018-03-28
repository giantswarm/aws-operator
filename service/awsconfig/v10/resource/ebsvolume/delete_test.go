package ebsvolume

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_newDeleteChange(t *testing.T) {
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
		currentState  *EBSVolumeState
		desiredState  *EBSVolumeState
		expectedState *EBSVolumeState
	}{
		{
			description: "case 0: basic match",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						VolumeID: "vol-1234",
					},
					{
						VolumeID: "vol-5678",
					},
				},
			},
			desiredState: nil,
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						VolumeID: "vol-1234",
					},
					{
						VolumeID: "vol-5678",
					},
				},
			},
		},
		{
			description: "case 1: basic match with attachments",
			obj:         customObject,
			currentState: &EBSVolumeState{
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
			desiredState: nil,
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
		},
		{
			description:   "case 2: return nil when current state is nil",
			obj:           customObject,
			currentState:  nil,
			desiredState:  nil,
			expectedState: nil,
		},
		{
			description: "case 3: return nil when current volumes are empty",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []ebs.Volume{},
			},
			desiredState:  nil,
			expectedState: nil,
		},
		{
			description: "case 4: return nil when desired state is not nil",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						VolumeID: "vol-1234",
					},
				},
			},
			desiredState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						VolumeID: "vol-1234",
					},
				},
			},
			expectedState: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ebsConfig := ebs.Config{
				Client: &ebs.EC2ClientMock{},
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

			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			deleteState, ok := result.(*EBSVolumeState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", deleteState, result)
			}

			if !reflect.DeepEqual(deleteState, tc.expectedState) {
				t.Errorf("expected delete state '%#v', got '%#v'", tc.expectedState, deleteState)
			}
		})
	}
}
