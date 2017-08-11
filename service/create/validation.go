package create

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/microerror"
)

func validateMasters(awsMasters []aws.Node, masters []spec.Node) error {
	// Currently only a single master is expected.
	if len(awsMasters) != 1 || len(masters) != 1 {
		return microerror.Mask(invalidMasterNodeCountError)
	}

	return nil
}

func validateWorkers(awsWorkers []aws.Node, workers []spec.Node) error {
	if len(awsWorkers) < 1 || len(workers) < 1 {
		return microerror.Mask(workersListEmptyError)
	}

	if len(awsWorkers) != len(workers) {
		return microerror.Mask(invalidWorkerNodeCountError)
	}

	firstImageID := awsWorkers[0].ImageID
	firstInstanceType := awsWorkers[0].InstanceType
	for _, worker := range awsWorkers {
		if worker.ImageID != firstImageID {
			return microerror.Mask(differentImageIDsError)
		}
		if worker.InstanceType != firstInstanceType {
			return microerror.Mask(differentInstanceTypesError)
		}
	}

	return nil
}

func validateCluster(cluster awstpr.CustomObject) error {
	if err := validateMasters(cluster.Spec.AWS.Masters, cluster.Spec.Cluster.Masters); err != nil {
		return microerror.Mask(err)
	}

	if err := validateWorkers(cluster.Spec.AWS.Workers, cluster.Spec.Cluster.Workers); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
