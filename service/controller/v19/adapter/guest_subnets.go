package adapter

import (
	"fmt"
	"sort"

	"github.com/giantswarm/aws-operator/service/controller/v19/key"
)

type Subnet struct {
	Index               int
	AvailabilityZone    string
	CIDR                string
	Name                string
	MapPublicIPOnLaunch bool
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
	for i, az := range zones {
		snet := Subnet{
			Index:               i,
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Public.CIDR,
			Name:                fmt.Sprintf("PublicSubnet%d", i),
			MapPublicIPOnLaunch: false,
		}
		s.PublicSubnets = append(s.PublicSubnets, snet)

		snet = Subnet{
			Index:               i,
			AvailabilityZone:    az.Name,
			CIDR:                az.Subnet.Private.CIDR,
			Name:                fmt.Sprintf("PrivateSubnet%d", i),
			MapPublicIPOnLaunch: false,
		}
		s.PrivateSubnets = append(s.PrivateSubnets, snet)
	}

	return nil
}
