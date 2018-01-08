package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type InternetGateway struct {
	Name      string
	VpcID     string
	id        string
	// Dependencies.
	Logger micrologger.Logger
	AWSEntity
}

func (g InternetGateway) findExisting() (*ec2.InternetGateway, error) {
	gateways, err := g.Clients.EC2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
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

	if len(gateways.InternetGateways) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, InternetGatewayType, g.Name)
	} else if len(gateways.InternetGateways) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return gateways.InternetGateways[0], nil
}

func (g *InternetGateway) checkIfExists() (bool, error) {
	_, err := g.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (g *InternetGateway) CreateIfNotExists() (bool, error) {
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

func (g *InternetGateway) CreateOrFail() error {
	gateway, err := g.Clients.EC2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return microerror.Mask(err)
	}
	gatewayID := *gateway.InternetGateway.InternetGatewayId

	if _, err := g.Clients.EC2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(gatewayID),
		VpcId:             aws.String(g.VpcID),
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := g.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(gatewayID),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(g.Name),
			},
			{
				Key:   aws.String(tagKeyClusterId),
				Value: aws.String(g.Name),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	g.id = gatewayID

	return nil
}

func (g *InternetGateway) Delete() error {
	gateway, err := g.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	detachOperation := func() error {
		if _, err := g.Clients.EC2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: gateway.InternetGatewayId,
			VpcId:             aws.String(g.VpcID),
		}); err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	detachNotify := NewNotify(g.Logger, "detaching internet gateway")
	if err := backoff.RetryNotify(detachOperation, NewCustomExponentialBackoff(), detachNotify); err != nil {
		return microerror.Mask(err)
	}

	deleteOperation := func() error {
		if _, err := g.Clients.EC2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: gateway.InternetGatewayId,
		}); err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	deleteNotify := NewNotify(g.Logger, "deleting internet gateway")
	if err := backoff.RetryNotify(deleteOperation, NewCustomExponentialBackoff(), deleteNotify); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (g InternetGateway) GetID() (string, error) {
	if g.id != "" {
		return g.id, nil
	}

	gateway, err := g.findExisting()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *gateway.InternetGatewayId, nil
}
