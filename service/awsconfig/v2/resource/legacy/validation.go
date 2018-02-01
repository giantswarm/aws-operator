package legacy

import (
	"regexp"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	availabilityZoneFormat = "[\\d][a-z]"
	// Maximum AWS idle timeout is 60 minutes
	maxIdleTimeout = 60 * 60
)

func validateAvailabilityZone(cluster v1alpha1.AWSConfig) error {
	az := cluster.Spec.AWS.AZ
	region := cluster.Spec.AWS.Region

	// AZ should begin with the Region name.
	if !strings.HasPrefix(az, region) {
		return microerror.Mask(invalidAvailabilityZoneError)
	}
	// AZ format uses 2 hyphens as a separator.
	if len(strings.Split(az, "-")) != 3 {
		return microerror.Mask(invalidAvailabilityZoneError)
	}

	// Check format of the AZ suffix.
	regEx, err := regexp.Compile(availabilityZoneFormat)
	if err != nil {
		return microerror.Mask(err)
	}
	if !regEx.MatchString(strings.Split(az, "-")[2]) {
		return microerror.Mask(invalidAvailabilityZoneError)
	}

	return nil
}

func validateELB(aws v1alpha1.AWSConfigSpecAWS) error {
	if aws.API.ELB.IdleTimeoutSeconds > maxIdleTimeout {
		return microerror.Maskf(idleTimeoutSecondsOutOfRangeError, idleTimeoutSecondsOutOfRangeErrorFormat, "api")
	}
	if aws.Etcd.ELB.IdleTimeoutSeconds > maxIdleTimeout {
		return microerror.Maskf(idleTimeoutSecondsOutOfRangeError, idleTimeoutSecondsOutOfRangeErrorFormat, "etcd")
	}
	if aws.Ingress.ELB.IdleTimeoutSeconds > maxIdleTimeout {
		return microerror.Maskf(idleTimeoutSecondsOutOfRangeError, idleTimeoutSecondsOutOfRangeErrorFormat, "ingress")
	}

	return nil
}

func validateMasters(awsMasters []v1alpha1.AWSConfigSpecAWSNode, masters []v1alpha1.ClusterNode) error {
	// Currently only a single master is expected.
	if len(awsMasters) != 1 || len(masters) != 1 {
		return microerror.Mask(invalidMasterNodeCountError)
	}

	return nil
}

func validateWorkers(awsWorkers []v1alpha1.AWSConfigSpecAWSNode, workers []v1alpha1.ClusterNode) error {
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

func validateCluster(cluster v1alpha1.AWSConfig) error {
	if err := validateAvailabilityZone(cluster); err != nil {
		return microerror.Mask(err)
	}

	if err := validateELB(cluster.Spec.AWS); err != nil {
		return microerror.Mask(err)
	}

	if err := validateMasters(cluster.Spec.AWS.Masters, cluster.Spec.Cluster.Masters); err != nil {
		return microerror.Mask(err)
	}

	if err := validateWorkers(cluster.Spec.AWS.Workers, cluster.Spec.Cluster.Workers); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
