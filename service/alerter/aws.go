package alerter

import (
	"fmt"

	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/microerror"
)

// ListMasterInstances lists the master instances for this cluster.
func (s *Service) ListMasterInstances(clusterID string) ([]string, error) {
	instanceNames := []string{}

	instances, err := awsresources.FindInstances(awsresources.FindInstancesInput{
		Clients: s.awsClients,
		Logger:  s.logger,
		Pattern: fmt.Sprintf("%s-master-0", clusterID),
	})
	if err != nil {
		return instanceNames, microerror.Mask(err)
	}

	for _, instance := range instances {
		instanceNames = append(instanceNames, instance.Name)
	}

	return instanceNames, nil
}

// ListVpcs lists the VPCs in this installation and returns the clusterIDs
// associated with them.
func (s *Service) ListVpcs() ([]string, error) {
	clusterIDs := []string{}

	vpc := &awsresources.VPC{
		InstallationName: s.installationName,
		AWSEntity:        awsresources.AWSEntity{Clients: s.awsClients},
	}

	vpcs, err := vpc.List()
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	for _, vpc := range vpcs {
		clusterIDs = append(clusterIDs, vpc.Name)
	}

	return clusterIDs, nil
}
