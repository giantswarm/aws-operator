package loadbalancer

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_CurrentState(t *testing.T) {
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
		expectedState LoadBalancerState
		loadBalancers []LoadBalancerMock
	}{
		{
			description: "basic match with no load balancers",
			obj:         customObject,
			expectedState: LoadBalancerState{
				LoadBalancerNames: []string{},
			},
		},
		{
			description: "basic match with load balancer",
			obj:         customObject,
			expectedState: LoadBalancerState{
				LoadBalancerNames: []string{
					"test-elb",
				},
			},
			loadBalancers: []LoadBalancerMock{
				{
					loadBalancerName: "test-elb",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&elb.Tag{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world"),
						},
					},
				},
			},
		},
	}
	var err error
	var newResource *Resource

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Clients: Clients{
					ELB: &ELBClientMock{
						loadBalancers: tc.loadBalancers,
					},
				},
				Logger: microloggertest.New(),
			}
			newResource, err = New(c)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			result, err := newResource.GetCurrentState(context.TODO(), tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			currentState, ok := result.(LoadBalancerState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if !reflect.DeepEqual(currentState, tc.expectedState) {
				t.Errorf("expected current state '%#v', got '%#v'", tc.expectedState, currentState)
			}

		})
	}
}
