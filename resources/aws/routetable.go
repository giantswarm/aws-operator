package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

type RouteTable struct {
	Name  string
	VpcID string
	id    string
	AWSEntity
}

func (r RouteTable) findExisting() (*ec2.RouteTable, error) {
	routeTables, err := r.Clients.EC2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(r.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(routeTables.RouteTables) < 1 {
		return nil, microerror.MaskAny(routeTableFindError)
	}

	return routeTables.RouteTables[0], nil
}

func (r *RouteTable) checkIfExists() (bool, error) {
	routeTable, err := r.findExisting()
	if err != nil {
		if strings.Contains(err.Error(), routeTableFindError.Error()) {
			return false, nil
		}
		return false, microerror.MaskAny(err)
	}

	r.id = *routeTable.RouteTableId

	return true, nil
}

func (r *RouteTable) CreateIfNotExists() (bool, error) {
	exists, err := r.checkIfExists()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if exists {
		return false, nil
	}

	if err := r.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (r *RouteTable) CreateOrFail() error {
	routeTable, err := r.Clients.EC2.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: aws.String(r.VpcID),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := r.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{routeTable.RouteTable.RouteTableId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(r.Name),
			},
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	r.id = *routeTable.RouteTable.RouteTableId

	return nil
}

func (r *RouteTable) Delete() error {
	routeTable, err := r.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := r.Clients.EC2.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: routeTable.RouteTableId,
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (r RouteTable) ID() string {
	return r.id
}
