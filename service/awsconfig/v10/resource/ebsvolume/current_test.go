package ebsvolume

import (
	"context"
	"reflect"
	"testing"

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
		ebsVolumes    []ebs.Volume
		expectedState *EBSVolumeState
	}{
		{
			description: "case 0: basic match with no ebs volumes",
			obj:         customObject,
			ebsVolumes:  []ebs.Volume{},
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{},
			},
		},
		{
			description: "case 1: basic match with ebs volume",
			obj:         customObject,
			ebsVolumes: []ebs.Volume{
				{
					Attachments: []ebs.VolumeAttachment{},
					VolumeID:    "vol-1234",
				},
			},
			expectedState: &EBSVolumeState{
				Volumes: []ebs.Volume{
					{
						Attachments: []ebs.VolumeAttachment{},
						VolumeID:    "vol-1234",
					},
				},
			},
		},
		{
			description: "case 2: basic match with attachment",
			obj:         customObject,
			ebsVolumes: []ebs.Volume{
				{
					Attachments: []ebs.VolumeAttachment{
						{
							InstanceID: "i-12345",
							Device:     "/dev/sdh",
						},
					},
					VolumeID: "vol-1234",
				},
			},
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
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Logger: microloggertest.New(),
				Service: &EBSServiceMock{
					volumes: tc.ebsVolumes,
				},
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
