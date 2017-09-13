package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

type Route struct {
	DestinationCidrBlock string
	VpcID                string
	RouteTable           RouteTable
	AWSEntity
}

func (r Route) findExisting() (*ec2.Route, error) {

	awsRouteTable, err := r.RouteTable.findExisting()

	for _, route := range awsRouteTable.Routes {
		if route.DestinationCidrBlock != nil && route.VpcPeeringConnectionId != nil &&
			*route.VpcPeeringConnectionId == r.VpcID && *route.DestinationCidrBlock == r.DestinationCidrBlock {
			return route, nil
		}
	}
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, RouteType)
}

func (r *Route) checkIfExists() (bool, error) {
	_, err := r.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (r *Route) CreateIfNotExists() (bool, error) {
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

func (r *Route) CreateOrFail() error {
	routeTableID, err := r.RouteTable.GetID()
	if err != nil {
		return microerror.Mask(err)
	}
	if _, err := r.Clients.EC2.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:           &routeTableID,
		DestinationCidrBlock:   &r.DestinationCidrBlock,
		VpcPeeringConnectionId: &r.VpcID,
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Route) Delete() error {
	routeTableID, err := r.RouteTable.GetID()
	if err != nil {
		return microerror.Mask(err)
	}
	if _, err := r.Clients.EC2.DeleteRoute(&ec2.DeleteRouteInput{
		RouteTableId:         &routeTableID,
		DestinationCidrBlock: &r.DestinationCidrBlock,
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
