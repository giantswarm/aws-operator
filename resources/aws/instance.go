package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	microerror "github.com/giantswarm/microkit/error"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

const (
	// EC2 instance tag keys.
	tagKeyName    string = "Name"
	tagKeyCluster string = "Cluster"
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
	InstanceID             string
	KeyName                string
	MinCount               int
	MaxCount               int
	UserData               string
	SmallCloudconfig       string
	IamInstanceProfileName string
	PlacementAZ            string
	id                     string
	AWSEntity
}

func allExistingInstancesMatch(instances *ec2.DescribeInstancesOutput, state EC2StateCode) bool {
	// If the instance doesn't exist, then the Reservation field should be nil.
	// Otherwise, it will contain a slice of instances (which is going to contain our one instance we queried for).
	// TODO(nhlfr): Check whether the instance has correct parameters. That will be most probably done when we
	// will introduce the interface for creating, deleting and updating resources.
	if instances.Reservations != nil {
		for _, r := range instances.Reservations {
			for _, i := range r.Instances {
				if *i.State.Code != int64(state) {
					return false
				}
			}
		}
	}
	return true
}

func existingInstanceID(instances *ec2.DescribeInstancesOutput) string {
	for _, r := range instances.Reservations {
		for _, i := range r.Instances {
			return *i.InstanceId
		}
	}
	return ""
}

func (i *Instance) CreateIfNotExists() (bool, error) {
	instances, err := i.Clients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(i.Name),
				},
			},
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyCluster)),
				Values: []*string{
					aws.String(i.ClusterName),
				},
			},
		},
	})
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if !allExistingInstancesMatch(instances, EC2TerminatedState) {
		i.InstanceID = existingInstanceID(instances)
		return false, nil
	}

	return true, i.CreateOrFail()
}

func (i *Instance) CreateOrFail() error {
	var reservation *ec2.Reservation
	reserveOperation := func() error {
		var err error
		reservation, err = i.Clients.EC2.RunInstances(&ec2.RunInstancesInput{
			ImageId:      aws.String(i.ImageID),
			InstanceType: aws.String(i.InstanceType),
			KeyName:      aws.String(i.KeyName),
			MinCount:     aws.Int64(int64(1)),
			MaxCount:     aws.Int64(int64(1)),
			UserData:     aws.String(i.SmallCloudconfig),
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: aws.String(i.IamInstanceProfileName),
			},
			Placement: &ec2.Placement{
				AvailabilityZone: aws.String(i.PlacementAZ),
			},
		})
		if err != nil {
			return microerror.MaskAny(err)
		}
		return nil
	}

	if err := backoff.Retry(reserveOperation, backoff.NewExponentialBackOff()); err != nil {
		return microerror.MaskAny(err)
	}

	for _, rawInstance := range reservation.Instances {
		i.InstanceID = *rawInstance.InstanceId

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
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (i *Instance) Delete() error {
	if _, err := i.Clients.EC2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(i.id),
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

type FindInstancesInput struct {
	Clients awsutil.Clients
	Pattern string
}

func FindInstances(input FindInstancesInput) ([]*Instance, error) {
	reservations, err := input.Clients.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(fmt.Sprintf("%s*", input.Pattern)),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	instances := make([]*Instance, 0, len(reservations.Reservations))

	for _, reservation := range reservations.Reservations {
		for _, rawInstance := range reservation.Instances {
			instances = append(instances, &Instance{
				id:        *rawInstance.InstanceId,
				AWSEntity: AWSEntity{Clients: input.Clients},
			})
		}
	}

	return instances, nil
}
