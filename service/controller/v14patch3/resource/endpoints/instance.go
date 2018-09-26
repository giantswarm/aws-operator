package endpoints

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch3/controllercontext"
)

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	ec2RunningState = 16
	tagKeyName      = "Name"
)

func (r Resource) findMasterInstance(ctx context.Context, instanceName string) (*ec2.Instance, error) {

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	filters := []*ec2.Filter{
		{
			Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
			Values: []*string{
				aws.String(instanceName),
			},
		},
	}

	output, err := sc.AWSClient.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
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
		return nil, microerror.Maskf(notFoundError, "instance: %s", instanceName)
	}
	if instancesFound > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "instances: %s", instanceName)
	}

	return masterInstance, nil
}
