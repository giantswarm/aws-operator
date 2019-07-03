package adapter

import (
	"sort"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
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
	zones, err := key.StatusAvailabilityZones(cfg.MachineDeployment)
	if err != nil {
		return microerror.Mask(err)
	}

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name
	})

	{
		numAZs := len(zones)
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	for _, az := range zones {
		snetName := key.SanitizeCFResourceName(key.PublicSubnetName(az.Name))
		snet := Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PublicSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.RouteTableName(cfg.CustomObject, suffixPublic, key.MasterAvailabilityZone(cfg.CustomObject))),
				SubnetName:     snetName,
			},
		}
		s.PublicSubnets = append(s.PublicSubnets, snet)

		snetName = key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name))
		snet = Subnet{
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR,
			Name:                snetName,
			MapPublicIPOnLaunch: false,
			RouteTableAssociation: RouteTableAssociation{
				Name:           key.SanitizeCFResourceName(key.PrivateSubnetRouteTableAssociationName(az.Name)),
				RouteTableName: key.SanitizeCFResourceName(key.PrivateRouteTableName(az.Name)),
				SubnetName:     snetName,
			},
		}
		s.PrivateSubnets = append(s.PrivateSubnets, snet)
	}

	return nil
}
