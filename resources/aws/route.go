package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
)

type Route struct {
	RouteTableID         string
	DestinationCidrBlock string
	VpcID                string
	AWSEntity
}

func (r *Route) CreateIfNotExists() (bool, error) {
	return false, microerror.Mask(notImplementedMethodError)
}

func (r *Route) CreateOrFail() error {

	if _, err := r.Clients.EC2.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:           &r.RouteTableID,
		DestinationCidrBlock:   &r.DestinationCidrBlock,
		VpcPeeringConnectionId: &r.VpcID,
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Route) Delete() error {

	if _, err := r.Clients.EC2.DeleteRoute(&ec2.DeleteRouteInput{
		RouteTableId:         &r.RouteTableID,
		DestinationCidrBlock: &r.DestinationCidrBlock,
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
