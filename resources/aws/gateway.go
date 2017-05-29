package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
)

type Gateway struct {
	Name  string
	VpcID string
	id    string
	// Dependencies.
	Logger micrologger.Logger
	AWSEntity
}

func (g Gateway) findExisting() (*ec2.InternetGateway, error) {
	gateways, err := g.Clients.EC2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(g.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(gateways.InternetGateways) < 1 {
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, GatewayType, g.Name)
	} else if len(gateways.InternetGateways) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return gateways.InternetGateways[0], nil
}

func (g *Gateway) checkIfExists() (bool, error) {
	_, err := g.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (g *Gateway) CreateIfNotExists() (bool, error) {
	exists, err := g.checkIfExists()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if exists {
		return false, nil
	}

	if err := g.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (g *Gateway) CreateOrFail() error {
	gateway, err := g.Clients.EC2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return microerror.MaskAny(err)
	}
	gatewayID := *gateway.InternetGateway.InternetGatewayId

	if _, err := g.Clients.EC2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(gatewayID),
		VpcId:             aws.String(g.VpcID),
	}); err != nil {
		return microerror.MaskAny(err)
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
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	g.id = gatewayID

	return nil
}

func (g *Gateway) Delete() error {
	gateway, err := g.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	detachOperation := func() error {
		if _, err := g.Clients.EC2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: gateway.InternetGatewayId,
			VpcId:             aws.String(g.VpcID),
		}); err != nil {
			return microerror.MaskAny(err)
		}
		return nil
	}
	detachNotify := NewNotify(g.Logger, "detaching gateway")
	if err := backoff.RetryNotify(detachOperation, NewCustomExponentialBackoff(), detachNotify); err != nil {
		return microerror.MaskAny(err)
	}

	deleteOperation := func() error {
		if _, err := g.Clients.EC2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: gateway.InternetGatewayId,
		}); err != nil {
			return microerror.MaskAny(err)
		}
		return nil
	}
	deleteNotify := NewNotify(g.Logger, "deleting gateway")
	if err := backoff.RetryNotify(deleteOperation, NewCustomExponentialBackoff(), deleteNotify); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (g Gateway) GetID() (string, error) {
	if g.id != "" {
		return g.id, nil
	}

	gateway, err := g.findExisting()
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	return *gateway.InternetGatewayId, nil
}
