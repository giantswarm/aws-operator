package loadbalancer

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v14/controllercontext"
)

func Test_clusterLoadBalancers(t *testing.T) {
	t.Parallel()
	customObject := v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		obj           v1alpha1.AWSConfig
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
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
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
						{
							Key:   aws.String("kubernetes.io/cluster/other-cluster"),
							Value: aws.String("owned"),
						},
						{
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
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world"),
						},
					},
				},
				{
					loadBalancerName: "test-elb-2",
					loadBalancerTags: []*elb.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
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
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world"),
						},
					},
				},
				{
					loadBalancerName: "test-elb-2",
					loadBalancerTags: []*elb.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/test-cluster"),
							Value: aws.String("owned"),
						},
						{
							Key:   aws.String("kubernetes.io/service-name"),
							Value: aws.String("hello-world-2"),
						},
					},
				},
				{
					loadBalancerName: "test-elb-3",
					loadBalancerTags: []*elb.Tag{
						{
							Key:   aws.String("kubernetes.io/cluster/another-cluster"),
							Value: aws.String("owned"),
						},
						{
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
						{
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
				Logger: microloggertest.New(),
			}
			newResource, err = New(c)
			if err != nil {
				t.Error("expected", nil, "got", err)
			}

			awsClients := awsclient.Clients{
				ELB: &ELBClientMock{
					loadBalancers: tc.loadBalancers,
				},
			}
			ctx := context.TODO()
			ctx = controllercontext.NewContext(ctx, controllercontext.Context{AWSClient: awsClients})

			result, err := newResource.clusterLoadBalancers(ctx, tc.obj)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !reflect.DeepEqual(result, tc.expectedState) {
				t.Errorf("expected current state '%#v', got '%#v'", tc.expectedState, result)
			}
		})
	}
}

func Test_splitLoadBalancers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		loadBalancerNames []*string
		chunkSize         int
		expectedChunks    [][]*string
	}{
		{
			name:              "case 0: empty lb names returns empty chunks",
			loadBalancerNames: []*string{},
			chunkSize:         20,
			expectedChunks:    [][]*string{},
		},
		{
			name: "case 1: single batch",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
			},
			chunkSize: 20,
			expectedChunks: [][]*string{{
				aws.String("lb-1"),
				aws.String("lb-2"),
			}},
		},
		{
			name: "case 2: multiple even chunks",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
				aws.String("lb-3"),
				aws.String("lb-4"),
				aws.String("lb-5"),
				aws.String("lb-6"),
			},
			chunkSize: 2,
			expectedChunks: [][]*string{
				{
					aws.String("lb-1"),
					aws.String("lb-2"),
				},
				{
					aws.String("lb-3"),
					aws.String("lb-4"),
				},
				{
					aws.String("lb-5"),
					aws.String("lb-6"),
				},
			},
		},
		{
			name: "case 3: multiple chunks of different sizes",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
				aws.String("lb-3"),
				aws.String("lb-4"),
				aws.String("lb-5"),
				aws.String("lb-6"),
				aws.String("lb-7"),
			},
			chunkSize: 3,
			expectedChunks: [][]*string{
				{
					aws.String("lb-1"),
					aws.String("lb-2"),
					aws.String("lb-3"),
				},
				{
					aws.String("lb-4"),
					aws.String("lb-5"),
					aws.String("lb-6"),
				},
				{
					aws.String("lb-7"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitLoadBalancers(tc.loadBalancerNames, tc.chunkSize)

			if !reflect.DeepEqual(result, tc.expectedChunks) {
				t.Fatalf("chunks == %q, want %q", result, tc.expectedChunks)
			}
		})
	}
}
