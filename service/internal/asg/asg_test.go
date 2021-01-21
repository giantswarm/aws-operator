package asg

import (
	"context"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-cmp/cmp"
)

func Test_ASG_drainable(t *testing.T) {
	testCases := []struct {
		name         string
		asgs         []*autoscaling.Group
		expectedName string
		errorMatcher func(err error) bool
	}{
		{
			name: "case 0",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "",
			errorMatcher: IsNoDrainable,
		},
		{
			name: "case 1",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "asg-2",
			errorMatcher: nil,
		},
		{
			name: "case 2",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "asg-1",
			errorMatcher: nil,
		},
		{
			name: "case 3",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "asg-1",
			errorMatcher: nil,
		},
		{
			name: "case 4",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingProceed),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "asg-1",
			errorMatcher: nil,
		},
		{
			name: "case 5",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingProceed),
						},
					},
				},
			},
			expectedName: "asg-3",
			errorMatcher: nil,
		},
		{
			name: "case 6",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-2"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("asg-3"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminating),
						},
					},
				},
			},
			expectedName: "",
			errorMatcher: IsNoDrainable,
		},
		{
			name: "case 7",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateInService),
						},
					},
				},
			},
			expectedName: "",
			errorMatcher: IsNoDrainable,
		},
		{
			name: "case 8",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingWait),
						},
					},
				},
			},
			expectedName: "asg-1",
			errorMatcher: nil,
		},
		{
			name: "case 9",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminatingProceed),
						},
					},
				},
			},
			expectedName: "asg-1",
			errorMatcher: nil,
		},
		{
			name: "case 10",
			asgs: []*autoscaling.Group{
				{
					AutoScalingGroupName: aws.String("asg-1"),
					Instances: []*autoscaling.Instance{
						{
							LifecycleState: aws.String(autoscaling.LifecycleStateTerminating),
						},
					},
				},
			},
			expectedName: "",
			errorMatcher: IsNoDrainable,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			name, err := drainable(context.Background(), tc.asgs)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !cmp.Equal(name, tc.expectedName) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expectedName, name))
			}
		})
	}
}

func Test_ASG_namesFromInstances(t *testing.T) {
	testCases := []struct {
		name      string
		instances []*ec2.Instance
		names     []string
	}{
		{
			name:      "case 0",
			instances: []*ec2.Instance{},
			names:     nil,
		},
		{
			name: "case 1",
			instances: []*ec2.Instance{
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(""),
							Value: aws.String(""),
						},
					},
				},
			},
			names: nil,
		},
		{
			name: "case 2",
			instances: []*ec2.Instance{
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-a"),
						},
					},
				},
			},
			names: []string{
				"asg-a",
			},
		},
		{
			name: "case 3",
			instances: []*ec2.Instance{
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-a"),
						},
					},
				},
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-b"),
						},
					},
				},
			},
			names: []string{
				"asg-a",
				"asg-b",
			},
		},
		{
			name: "case 4",
			instances: []*ec2.Instance{
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-a"),
						},
					},
				},
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(""),
							Value: aws.String(""),
						},
					},
				},
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("foo"),
							Value: aws.String("bar"),
						},
					},
				},
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-b"),
						},
					},
				},
				{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("aws:autoscaling:groupName"),
							Value: aws.String("asg-g"),
						},
					},
				},
			},
			names: []string{
				"asg-a",
				"asg-b",
				"asg-g",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			names := namesFromInstances(tc.instances)

			if !cmp.Equal(names, tc.names) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.names, names))
			}
		})
	}
}
