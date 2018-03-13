// +build k8srequired

package integration

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/client"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
)

var (
	f *framework.Framework
	c *client.AWS
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var v int
	var err error

	f, err = framework.New()
	if err != nil {
		panic(err.Error())
	}

	c = client.NewAWS()

	err = setupHostPeerVPC()
	if err != nil {
		panic(err.Error())
	}

	if err := f.SetUp(); err != nil {
		log.Printf("%v\n", err)
		v = 1
	}

	err = setup()
	if err != nil {
		log.Printf("%v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		f.DeleteGuestCluster()

		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown()
			if err != nil {
				log.Printf("%v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			f.TearDown()
		}

		err := teardownHostPeerVPC()
		if err != nil {
			log.Printf("%v\n", err)
			v = 1
		}
	}

	os.Exit(v)
}

func TestGuestReadyAfterMasterReboot(t *testing.T) {
	log.Println("getting master ID")
	describeInput := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(fmt.Sprintf("%s-master", ClusterID())),
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

func TestWorkersScaling(t *testing.T) {
	currentWorkers, err := numberOfWorkers(ClusterID())
	if err != nil {
		t.Fatalf("unexpected error getting number of workers %v", err)
	}
	currentMasters, err := numberOfMasters(ClusterID())
	if err != nil {
		t.Fatalf("unexpected error getting number of masters %v", err)
	}

	// increase number of workers
	expectedWorkers := currentWorkers + 1
	log.Printf("Increasing the number of workers to %d", expectedWorkers)
	err = addWorker(ClusterID())
	if err != nil {
		t.Fatalf("unexpected error setting number of workers to %d, %v", expectedWorkers, err)
	}

	if err := f.WaitForNodesUp(currentMasters + expectedWorkers); err != nil {
		t.Fatalf("unexpected error waiting for %d nodes up, %v", expectedWorkers, err)
	}
	log.Printf("%d worker nodes ready", expectedWorkers)

	// decrease number of workers
	expectedWorkers--
	log.Printf("Decreasing the number of workers to %d", expectedWorkers)
	err = removeWorker(ClusterID())
	if err != nil {
		t.Fatalf("unexpected error setting number of workers to %d, %v", expectedWorkers, err)
	}

	if err := f.WaitForNodesUp(currentMasters + expectedWorkers); err != nil {
		t.Fatalf("unexpected error waiting for %d nodes up, %v", expectedWorkers, err)
	}
	log.Printf("%d worker nodes ready", expectedWorkers)
}

func numberOfWorkers(clusterName string) (int, error) {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(*cluster), nil
}

func numberOfMasters(clusterName string) (int, error) {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.MasterCount(*cluster), nil
}

func addWorker(clusterName string) error {
	cluster, err := f.AWSCluster(clusterName)
	if err != nil {
		return microerror.Mask(err)
	}

	newWorker := cluster.Spec.AWS.Workers[0]

	patch := make([]framework.PatchSpec, 1)
	patch[0].Op = "add"
	patch[0].Path = "/spec/aws/workers/-"
	patch[0].Value = newWorker

	return f.ApplyAWSConfigPatch(patch, clusterName)
}

func removeWorker(clusterName string) error {
	patch := make([]framework.PatchSpec, 1)
	patch[0].Op = "remove"
	patch[0].Path = "/spec/aws/workers/1"

	return f.ApplyAWSConfigPatch(patch, clusterName)
}
