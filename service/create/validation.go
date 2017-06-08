package create

import (
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/awstpr/aws"
	microerror "github.com/giantswarm/microkit/error"
)

func validateWorkers(workers []aws.Node) error {
	if len(workers) < 1 {
		return microerror.MaskAny(workersListEmptyError)
	}

	firstImageID := workers[0].ImageID
	firstInstanceType := workers[0].InstanceType
	for _, worker := range workers {
		if worker.ImageID != firstImageID {
			return microerror.MaskAny(differentImageIDsError)
		}
		if worker.InstanceType != firstInstanceType {
			return microerror.MaskAny(differentInstanceTypesError)
		}
	}

	return nil
}

func validateCluster(cluster awstpr.CustomObject) error {
	if err := validateWorkers(cluster.Spec.AWS.Workers); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
