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
		expectedState *LoadBalancerState
		loadBalancers []LoadBalancerMock
	}{
		{
			description: "basic match with no load balancers",
			obj:         customObject,
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{},
			},
		},
		{
			description: "basic match with load balancer",
			obj:         customObject,
			expectedState: &LoadBalancerState{
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
		{
			description: "no matching load balancer",
			obj:         customObject,
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{},
			},
			loadBalancers: []LoadBalancerMock{
				{
					loadBalancerName: "test-elb",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/other-cluster"),
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
		{
			description: "multiple load balancers",
			obj:         customObject,
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{
					"test-elb",
					"test-elb-2",
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
				{
					loadBalancerName: "test-elb-2",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&elb.Tag{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world-2"),
						},
					},
				},
			},
		},
		{
			description: "multiple load balancers some not matching",
			obj:         customObject,
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{
					"test-elb",
					"test-elb-2",
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
				{
					loadBalancerName: "test-elb-2",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						&elb.Tag{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world-2"),
						},
					},
				},
				{
					loadBalancerName: "test-elb-3",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/another-cluster"),
							Value: aws.String("owned"),
						},
						&elb.Tag{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world-2"),
						},
					},
				},
			},
		},
		{
			description: "missing service tag",
			obj:         customObject,
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{},
			},
			loadBalancers: []LoadBalancerMock{
				{
					loadBalancerName: "test-elb",
					loadBalancerTags: []*elb.Tag{
						&elb.Tag{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
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
			currentState, ok := result.(*LoadBalancerState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", currentState, result)
			}

			if !reflect.DeepEqual(currentState, tc.expectedState) {
				t.Errorf("expected current state '%#v', got '%#v'", tc.expectedState, currentState)
			}
		})
	}
}
