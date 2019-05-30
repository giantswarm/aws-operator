package adapter

import (
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

type Subnet struct {
	AvailabilityZone      string
	CIDR                  string
	Name                  string
	MapPublicIPOnLaunch   bool
	RouteTableAssociation RouteTableAssociation
}

type RouteTableAssociation struct {
	Name           string
	RouteTableName string
	SubnetName     string
}

type GuestSubnetsAdapter struct {
	PublicSubnets  []Subnet
	PrivateSubnets []Subnet
}

func (s *GuestSubnetsAdapter) Adapt(cfg Config) error {
	zones := key.StatusAvailabilityZones(cfg.MachineDeployment)
	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	{
		numAZs := len(zones)
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	for i, az := range zones {
		snetName := key.PublicSubnetName(i)
		snet := Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.PublicSubnetRouteTableAssociationName(i),
				RouteTableName: "PublicRouteTable",
				SubnetName:     snetName,
			},
		}
		s.PublicSubnets = append(s.PublicSubnets, snet)

		snetName = key.PrivateSubnetName(i)
		snet = Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.PrivateSubnetRouteTableAssociationName(i),
				RouteTableName: key.PrivateRouteTableName(i),
				SubnetName:     snetName,
			},
		}
		s.PrivateSubnets = append(s.PrivateSubnets, snet)
	}

	return nil
}
