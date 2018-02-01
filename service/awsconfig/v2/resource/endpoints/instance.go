package endpoints

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	ec2RunningState = 16
	tagKeyName      = "Name"
)

func (r Resource) findMasterInstance(instanceName string) (*ec2.Instance, error) {
	filters := []*ec2.Filter{
		{
			Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
			Values: []*string{
				aws.String(instanceName),
			},
		},
	}

	output, err := r.awsClients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var masterInstance *ec2.Instance
	var instancesFound int
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Code == ec2RunningState {
				masterInstance = instance
				instancesFound++
			}
		}
	}

	if instancesFound < 1 {
		return nil, microerror.Mask(notFoundError)
	}
	if instancesFound > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return masterInstance, nil
}
