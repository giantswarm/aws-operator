package create

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/microerror"
)

func validateWorkers(workers []aws.Node) error {
	if len(workers) < 1 {
		return microerror.Mask(workersListEmptyError)
	}

	firstImageID := workers[0].ImageID
	firstInstanceType := workers[0].InstanceType
	for _, worker := range workers {
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
	if err := validateWorkers(cluster.Spec.AWS.Workers); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
