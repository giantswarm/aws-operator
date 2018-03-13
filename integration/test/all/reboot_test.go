// +build k8srequired

package all

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/giantswarm/aws-operator/integration/env"
)

func TestGuestReadyAfterMasterReboot(t *testing.T) {
	log.Println("getting master ID")
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(fmt.Sprintf("%s-master", env.ClusterID())),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	}
	res, err := c.EC2.DescribeInstances(describeInput)
	if err != nil {
		t.Errorf("unexpected error getting master id %v", err)
	}
	if len(res.Reservations) != 1 {
		t.Errorf("unexpected number of reservations %d", len(res.Reservations))
	}
	if len(res.Reservations[0].Instances) != 1 {
		t.Errorf("unexpected number of instances %d", len(res.Reservations[0].Instances))
	}
	masterID := res.Reservations[0].Instances[0].InstanceId

	log.Println("rebooting master")
	rebootInput := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			masterID,
		},
	}
	_, err = c.EC2.RebootInstances(rebootInput)
	if err != nil {
		t.Errorf("unexpected error rebooting  master %v", err)
	}

	if err := f.WaitForAPIDown(); err != nil {
		t.Errorf("unexpected error waiting for master shutting down %v", err)
	}

	if err := f.WaitForGuestReady(); err != nil {
		t.Errorf("unexpected error waiting for guest cluster ready, %v", err)
	}
}
