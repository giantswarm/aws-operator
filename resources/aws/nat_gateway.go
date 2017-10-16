package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type NatGateway struct {
	Name   string
	Subnet *Subnet
	id     string
	// Dependencies.
	Logger micrologger.Logger
	AWSEntity
}

func (g NatGateway) findExisting() (*ec2.NatGateway, error) {
	gateways, err := g.Clients.EC2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(g.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(gateways.NatGateways) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, NatGatewayType, g.Name)
	} else if len(gateways.NatGateways) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return gateways.NatGateways[0], nil
}

func (g *NatGateway) checkIfExists() (bool, error) {
	_, err := g.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (g *NatGateway) CreateIfNotExists() (bool, error) {
	exists, err := g.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := g.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (g *NatGateway) CreateOrFail() error {
	eip, err := g.Clients.EC2.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	subnetID, err := g.Subnet.GetID()
	if err != nil {
		return microerror.Mask(err)
	}

	gateway, err := g.Clients.EC2.CreateNatGateway(&ec2.CreateNatGatewayInput{
		AllocationId: eip.AllocationId,
		SubnetId:     aws.String(subnetID),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	gatewayID := *gateway.NatGateway.NatGatewayId

	if _, err := g.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(gatewayID),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(g.Name),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	g.id = gatewayID

	return nil
}

func (g *NatGateway) Delete() error {
	gateway, err := g.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	allocationIDs := []string{}

	for _, gatewayAddress := range gateway.NatGatewayAddresses {
		if gatewayAddress.AllocationId != nil {
			allocationIDs = append(allocationIDs, *gatewayAddress.AllocationId)
		}
	}

	if _, err := g.Clients.EC2.DeleteTags(&ec2.DeleteTagsInput{
		Resources: []*string{
			gateway.NatGatewayId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(g.Name),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := g.Clients.EC2.DeleteNatGateway(&ec2.DeleteNatGatewayInput{
		NatGatewayId: gateway.NatGatewayId,
	}); err != nil {
		return microerror.Mask(err)
	}

	for _, allocationID := range allocationIDs {
		if _, err := g.Clients.EC2.ReleaseAddress(&ec2.ReleaseAddressInput{
			AllocationId: aws.String(allocationID),
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (g NatGateway) GetID() (string, error) {
	if g.id != "" {
		return g.id, nil
	}

	gateway, err := g.findExisting()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *gateway.NatGatewayId, nil
}
