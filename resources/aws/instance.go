package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type EC2StateCode int

const (
	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#InstanceState
	EC2PendingState      EC2StateCode = 0
	EC2RunningState      EC2StateCode = 16
	EC2ShuttingDownState EC2StateCode = 32
	EC2TerminatedState   EC2StateCode = 48
	EC2StoppingState     EC2StateCode = 64
	EC2StoppedState      EC2StateCode = 80
)

type Instance struct {
	Name                   string
	ClusterName            string
	ImageID                string
	InstanceType           string
	KeyName                string
	MinCount               int
	MaxCount               int
	UserData               string
	SmallCloudconfig       string
	IamInstanceProfileName string
	PlacementAZ            string
	SecurityGroupID        string
	SubnetID               string
	PrivateIpAddress       string
	id                     string
	// Dependencies.
	Logger micrologger.Logger
	AWSEntity
}

func statePendingOrRunning(instance *ec2.Instance) bool {
	stateCode := *instance.State.Code
	switch stateCode {
	case int64(EC2PendingState), int64(EC2RunningState):
		return true
	}

	return false
}

func stateTerminated(instance *ec2.Instance) bool {
	stateCode := *instance.State.Code
	switch stateCode {
	case int64(EC2TerminatedState):
		return true
	}

	return false
}

func (i Instance) findExisting() (*ec2.Instance, error) {
	filters := []*ec2.Filter{}
	if i.ClusterName != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String(fmt.Sprintf("tag:%s", tagKeyCluster)),
			Values: []*string{
				aws.String(i.ClusterName),
			},
		})
	}
	if i.Name != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
			Values: []*string{
				aws.String(i.Name),
			},
		})
	}
	if i.id != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("instance-id"),
			Values: []*string{
				aws.String(i.id),
			},
		})
	}

	reservations, err := i.Clients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var existingInstance *ec2.Instance
	var instancesFound int
	for _, reservation := range reservations.Reservations {
		for _, instance := range reservation.Instances {
			if !stateTerminated(instance) {
				existingInstance = instance
				instancesFound++
			}
		}
	}

	if instancesFound < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, InstanceType, i.Name)
	} else if instancesFound > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return existingInstance, nil
}

func (i *Instance) checkIfExists() (bool, error) {
	instance, err := i.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	i.id = *instance.InstanceId

	return true, nil
}

func (i *Instance) CreateIfNotExists() (bool, error) {
	exists, err := i.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := i.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (i *Instance) CreateOrFail() error {
	var reservation *ec2.Reservation
	reserveOperation := func() error {
		var err error

		instancesInput := &ec2.RunInstancesInput{
			ImageId:      aws.String(i.ImageID),
			InstanceType: aws.String(i.InstanceType),
			MinCount:     aws.Int64(int64(1)),
			MaxCount:     aws.Int64(int64(1)),
			UserData:     aws.String(i.SmallCloudconfig),
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: aws.String(i.IamInstanceProfileName),
			},
			Placement: &ec2.Placement{
				AvailabilityZone: aws.String(i.PlacementAZ),
			},
			SecurityGroupIds: []*string{
				aws.String(i.SecurityGroupID),
			},
			SubnetId: aws.String(i.SubnetID),
		}
		if i.KeyName != "" {
			instancesInput.KeyName = aws.String(i.KeyName)
		}

		reservation, err = i.Clients.EC2.RunInstances(instancesInput)
		if err != nil {

			return microerror.Mask(err)
		}
		return nil
	}
	reserveNotify := NewNotify(i.Logger, "creating instance")
	if err := backoff.RetryNotify(reserveOperation, NewCustomExponentialBackoff(), reserveNotify); err != nil {
		return microerror.Mask(err)
	}

	for _, rawInstance := range reservation.Instances {
		i.id = *rawInstance.InstanceId

		if _, err := i.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{rawInstance.InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(tagKeyName),
					Value: aws.String(i.Name),
				},
				{
					Key:   aws.String(tagKeyCluster),
					Value: aws.String(i.ClusterName),
				},
			},
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (i *Instance) Delete() error {
	instance, err := i.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := i.Clients.EC2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	if err := i.Clients.EC2.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (i Instance) ID() string {
	return i.id
}

type FindInstancesInput struct {
	Clients awsutil.Clients
	Logger  micrologger.Logger
	Pattern string
}

func FindInstances(input FindInstancesInput) ([]*Instance, error) {
	reservations, err := input.Clients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(fmt.Sprintf("%s*", input.Pattern)),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	instances := make([]*Instance, 0, len(reservations.Reservations))

	for _, reservation := range reservations.Reservations {
		for _, rawInstance := range reservation.Instances {
			if !statePendingOrRunning(rawInstance) {
				continue
			}
			instances = append(instances, &Instance{
				id:               *rawInstance.InstanceId,
				PrivateIpAddress: *rawInstance.PrivateIpAddress,
				// Dependencies.
				Logger:    input.Logger,
				AWSEntity: AWSEntity{Clients: input.Clients},
			})
		}
	}

	return instances, nil
}
