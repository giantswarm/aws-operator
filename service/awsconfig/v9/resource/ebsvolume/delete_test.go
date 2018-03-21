package ebsvolume

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
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
			description: "basic match",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []Volume{
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
				Volumes: []Volume{
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
			description: "basic match with attachments",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []Volume{
					{
						Attachments: []VolumeAttachment{
							{
								InstanceID: "i-12345",
								Device:     "/dev/sdh",
							},
						},
						VolumeID: "vol-1234",
					},
					{
						Attachments: []VolumeAttachment{
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
				Volumes: []Volume{
					{
						Attachments: []VolumeAttachment{
							{
								InstanceID: "i-12345",
								Device:     "/dev/sdh",
							},
						},
						VolumeID: "vol-1234",
					},
					{
						Attachments: []VolumeAttachment{
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
			description:   "return nil when current state is nil",
			obj:           customObject,
			currentState:  nil,
			desiredState:  nil,
			expectedState: nil,
		},
		{
			description: "return nil when current volumes are empty",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []Volume{},
			},
			desiredState:  nil,
			expectedState: nil,
		},
		{
			description: "return nil when desired state is not nil",
			obj:         customObject,
			currentState: &EBSVolumeState{
				Volumes: []Volume{
					{
						VolumeID: "vol-1234",
					},
				},
			},
			desiredState: &EBSVolumeState{
				Volumes: []Volume{
					{
						VolumeID: "vol-1234",
					},
				},
			},
			expectedState: nil,
		},
	}

	var err error
	var newResource *Resource

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Clients: Clients{
					EC2: &EC2ClientMock{},
				},
				Logger: microloggertest.New(),
			}
			newResource, err = New(c)
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
