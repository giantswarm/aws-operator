package adapter

import (
	"fmt"
	"sort"

	"github.com/giantswarm/aws-operator/service/controller/v19/key"
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
	zones := key.StatusAvailabilityZones(cfg.CustomObject)
	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	s.PublicSubnets = []Subnet{
		Subnet{
			AvailabilityZone:    zones[0].Name,
			CIDR:                zones[0].Subnet.Public.CIDR,
			Name:                "PublicSubnet",
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           "PublicRouteTableAssociation",
				RouteTableName: "PublicRouteTable",
				SubnetName:     "PublicSubnet",
			},
		},
	}

	s.PrivateSubnets = []Subnet{
		Subnet{
			AvailabilityZone:    zones[0].Name,
			CIDR:                zones[0].Subnet.Private.CIDR,
			Name:                "PrivateSubnet",
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           "PrivateRouteTableAssociation",
				RouteTableName: "PrivateRouteTable",
				SubnetName:     "PrivateSubnet",
			},
		},
	}

	for i := 1; i < len(zones); i++ {
		az := zones[i]
		snetName := fmt.Sprintf("PublicSubnet%02d", i)
		snet := Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           fmt.Sprintf("PublicRouteTableAssociation%02d", i),
				RouteTableName: "PublicRouteTable",
				SubnetName:     snetName,
			},
		}
		s.PublicSubnets = append(s.PublicSubnets, snet)

		snetName = fmt.Sprintf("PrivateSubnet%02d", i)
		snet = Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           fmt.Sprintf("PrivateRouteTableAssociation%02d", i),
				RouteTableName: fmt.Sprintf("PrivateRouteTable%02d", i),
				SubnetName:     snetName,
			},
		}
		s.PrivateSubnets = append(s.PrivateSubnets, snet)
	}

	return nil
}
