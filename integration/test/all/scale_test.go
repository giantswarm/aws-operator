// +build k8srequired

package all

import (
	"log"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/service/awsconfig/v9/key"
)

func TestWorkersScaling(t *testing.T) {
	currentWorkers, err := numberOfWorkers(env.ClusterID())
	if err != nil {
		t.Fatalf("unexpected error getting number of workers %v", err)
	}
	currentMasters, err := numberOfMasters(env.ClusterID())
	if err != nil {
		t.Fatalf("unexpected error getting number of masters %v", err)
	}

	// increase number of workers
	expectedWorkers := currentWorkers + 1
	log.Printf("Increasing the number of workers to %d", expectedWorkers)
	err = addWorker(env.ClusterID())
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
	err = removeWorker(env.ClusterID())
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
