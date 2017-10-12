package create

import (
	"fmt"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

type SubnetInput struct {
	CidrBlock  string
	Clients    awsutil.Clients
	Cluster    awstpr.CustomObject
	MakePublic bool
	Name       string
	RouteTable *awsresources.RouteTable
	VpcID      string
}

// createSubnet creates a subnet and optionally makes it public.
func (s *Service) createSubnet(input SubnetInput) (*awsresources.Subnet, error) {
	subnet := &awsresources.Subnet{
		AvailabilityZone: key.AvailabilityZone(input.Cluster),
		CidrBlock:        input.CidrBlock,
		Name:             input.Name,
		VpcID:            input.VpcID,
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

	if input.MakePublic && input.RouteTable != nil {
		err := subnet.MakePublic(input.RouteTable)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return subnet, nil
}

// deleteSubnet deletes a subnet.
func (s *Service) deleteSubnet(input SubnetInput) error {
	subnet := &awsresources.Subnet{
		Name: input.Name,
		// Dependencies.
		Logger:    s.logger,
		AWSEntity: awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := subnet.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete subnet '%s': '%#v'", input.Name, err))
	} else {
		s.logger.Log("info", "deleted subnet '%s'")
	}

	return nil
}
