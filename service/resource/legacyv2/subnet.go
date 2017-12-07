package legacyv2

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/keyv2"
)

type SubnetInput struct {
	CidrBlock  string
	Clients    awsutil.Clients
	Cluster    v1alpha1.AWSConfig
	MakePublic bool
	Name       string
	RouteTable *awsresources.RouteTable
	VpcID      string
}

// createSubnet creates a subnet and optionally makes it public.
func (s *Resource) createSubnet(input SubnetInput) (*awsresources.Subnet, error) {
	subnet := &awsresources.Subnet{
		AvailabilityZone: keyv2.AvailabilityZone(input.Cluster),
		CidrBlock:        input.CidrBlock,
		Name:             input.Name,
		VpcID:            input.VpcID,
		ClusterName:      keyv2.ClusterID(input.Cluster),
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: input.Clients},
	}
	subnetCreated, err := subnet.CreateIfNotExists()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if subnetCreated {
		s.logger.Log("info", fmt.Sprintf("created subnet '%s'", input.Name))
	} else {
		s.logger.Log("info", fmt.Sprintf("subnet '%s' already exists, reusing", input.Name))
	}

	err = subnet.AssociateRouteTable(input.RouteTable)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if input.MakePublic {
		err := subnet.MakePublic()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return subnet, nil
}

// deleteSubnet deletes a subnet.
func (s *Resource) deleteSubnet(input SubnetInput) error {
	subnet := &awsresources.Subnet{
		Name: input.Name,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := subnet.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete subnet '%s': '%#v'", input.Name, err))
	} else {
		s.logger.Log("info", fmt.Sprintf("deleted subnet '%s'", input.Name))
	}

	return nil
}
