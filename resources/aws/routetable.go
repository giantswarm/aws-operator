package aws

import (
	"fmt"

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

func (r RouteTable) list() ([]*ec2.RouteTable, error) {
	out, err := r.Clients.EC2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(r.Name),
				},
			},
		},
	})
	return out.RouteTables, microerror.MaskAny(err)
}

func (r *RouteTable) checkIfExists() (bool, error) {
	routeTables, err := r.list()
	if err == nil && len(routeTables) > 0 {
		r.id = *routeTables[0].RouteTableId
	}
	return len(routeTables) > 0, microerror.MaskAny(err)
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
	routeTables, err := r.list()
	if err != nil {
		return microerror.MaskAny(err)
	}

	for _, routeTable := range routeTables {
		_, err := r.Clients.EC2.DeleteRouteTable(&ec2.DeleteRouteTableInput{
			RouteTableId: routeTable.RouteTableId,
		})
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (r RouteTable) ID() string {
	return r.id
}
