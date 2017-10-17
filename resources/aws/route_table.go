package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type RouteTable struct {
	Name  string
	VpcID string
	id    string
	// Dependencies.
	Client *ec2.EC2
	Logger micrologger.Logger
}

func (r RouteTable) findExisting() (*ec2.RouteTable, error) {
	routeTables, err := r.Client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(r.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(routeTables.RouteTables) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, RouteTableType, r.Name)
	} else if len(routeTables.RouteTables) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return routeTables.RouteTables[0], nil
}

func (r *RouteTable) checkIfExists() (bool, error) {
	routeTable, err := r.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	r.id = *routeTable.RouteTableId

	return true, nil
}

func (r *RouteTable) CreateIfNotExists() (bool, error) {
	exists, err := r.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := r.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (r *RouteTable) CreateOrFail() error {
	routeTable, err := r.Client.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: aws.String(r.VpcID),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := r.Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{routeTable.RouteTable.RouteTableId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(r.Name),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	r.id = *routeTable.RouteTable.RouteTableId

	return nil
}

func (r *RouteTable) Delete() error {
	routeTable, err := r.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	for _, association := range routeTable.Associations {
		if _, err := r.Client.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
			AssociationId: association.RouteTableAssociationId,
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	if _, err := r.Client.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: routeTable.RouteTableId,
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r RouteTable) GetID() (string, error) {
	if r.id != "" {
		return r.id, nil
	}

	routeTable, err := r.findExisting()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *routeTable.RouteTableId, nil
}

// MakePublic creates a route that allows traffic from outside the VPC.
// To do that, it needs to add a route on the Internet Gateway of the VPC.
func (r RouteTable) MakePublic() error {
	gatewayID, err := r.getInternetGateway()
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := r.Client.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:         aws.String(r.id),
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(gatewayID),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// CreateNatGatewayRoute creates a default route to the NAT gateway for the
// private subnet. Retry is needed due to a delay while the gateway is created.
func (r RouteTable) CreateNatGatewayRoute(natGatewayID string) (bool, error) {
	createOperation := func() error {
		_, err := r.Client.CreateRoute(&ec2.CreateRouteInput{
			RouteTableId:         aws.String(r.id),
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			NatGatewayId:         aws.String(natGatewayID),
		})
		if err != nil {
			// Fall through if the rule already exists.
			if awsutil.IsRouteDuplicateError(err) {
				return nil
			}

			return microerror.Mask(err)
		}

		return nil
	}
	createNotify := NewNotify(r.Logger, "creating nat gateway route")
	if err := backoff.RetryNotify(createOperation, NewCustomExponentialBackoff(), createNotify); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

// getInternetGateway retrieves the Internet Gateway of the Route Table's VPC.
// An internet gateway is what enables communication between a VPC and the outside Intenet.
// See https://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Internet_Gateway.html
func (r RouteTable) getInternetGateway() (string, error) {
	resp, err := r.Client.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				// retrieve only the gateway attached to the vpc of the route table.
				Name: aws.String("attachment.vpc-id"),
				Values: []*string{
					aws.String(r.VpcID),
				},
			},
		},
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	if len(resp.InternetGateways) == 0 {
		return "", microerror.Maskf(notFoundError, notFoundErrorFormat, RouteTableType, r.Name)
	}

	return *resp.InternetGateways[0].InternetGatewayId, nil
}
