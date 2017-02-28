package create

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestAllExistingInstancesMatch(t *testing.T) {
	tests := []struct {
		name      string
		instances *ec2.DescribeInstancesOutput
		state     EC2StateCode
		res       bool
	}{
		{
			name: "Expect terminated with a single terminated instance in a single reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   true,
		},
		{
			name: "Expect terminated with three terminated instances in a single reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   true,
		},
		{
			name: "Expect not terminated with a terminated instance and a running instance in a single reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2RunningState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   false,
		},
		{
			name: "Expect not stopped with a single running instance in a single reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2RunningState)),
								},
							},
						},
					},
				},
			},
			state: EC2StoppedState,
			res:   false,
		},
		{
			name: "Expect running with a single running instance in a single reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2RunningState)),
								},
							},
						},
					},
				},
			},
			state: EC2RunningState,
			res:   true,
		},
		{
			name: "Expect terminated with two terminated instances in different reservations",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   true,
		},
		{
			name: "Expect not terminated with two terminated instances and one stopping instance in different reservations",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2StoppingState)),
								},
							},
						},
					},
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   false,
		},
		{
			name: "Expect not terminated with two terminated instances in one reservation, one terminated and one stopping in another reservation, and one terminated in another reservation",
			instances: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2StoppingState)),
								},
							},
						},
					},
					{
						Instances: []*ec2.Instance{
							{
								State: &ec2.InstanceState{
									Code: aws.Int64(int64(EC2TerminatedState)),
								},
							},
						},
					},
				},
			},
			state: EC2TerminatedState,
			res:   false,
		},
	}

	for _, tc := range tests {
		res := allExistingInstancesMatch(tc.instances, tc.state)
		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] Some instance didn't match the expected state", tc.name))
	}
}
