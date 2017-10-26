package alerter

import (
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/microerror"
)

// ListVpcs lists the VPCs in this installation and returns the clusterIDs
// associated with them.
func (s *Service) ListVpcs() ([]string, error) {
	clusterIDs := []string{}

	vpc := &awsresources.VPC{
		// TODO Make configurable.
		InstallationName: "gauss",
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
