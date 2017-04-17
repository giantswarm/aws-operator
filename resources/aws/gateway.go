package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

type Gateway struct {
	Name  string
	VpcID string
	id    string
	AWSEntity
}

func (g Gateway) list() ([]*ec2.InternetGateway, error) {
	out, err := g.Clients.EC2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(g.Name),
				},
			},
		},
	})
	return out.InternetGateways, microerror.MaskAny(err)
}

func (g *Gateway) checkIfExists() (bool, error) {
	gateways, err := g.list()
	if err == nil && len(gateways) > 0 {
		g.id = *gateways[0].InternetGatewayId
	}
	return len(gateways) > 0, microerror.MaskAny(err)
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
	gateways, err := g.list()
	if err != nil {
		return microerror.MaskAny(err)
	}

	for _, gateway := range gateways {
		_, err := g.Clients.EC2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: gateway.InternetGatewayId,
		})
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (g Gateway) ID() string {
	return g.id
}
