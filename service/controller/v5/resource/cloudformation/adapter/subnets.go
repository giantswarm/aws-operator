package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/v5/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v5/templates/cloudformation/guest/subnets.go
//

type subnetsAdapter struct {
	PublicSubnetAZ                   string
	PublicSubnetCIDR                 string
	PublicSubnetName                 string
	PublicSubnetMapPublicIPOnLaunch  bool
	PrivateSubnetAZ                  string
	PrivateSubnetCIDR                string
	PrivateSubnetName                string
	PrivateSubnetMapPublicIPOnLaunch bool
}

func (s *subnetsAdapter) getSubnets(cfg Config) error {
	s.PublicSubnetAZ = key.AvailabilityZone(cfg.CustomObject)
	s.PublicSubnetCIDR = cfg.CustomObject.Spec.AWS.VPC.PublicSubnetCIDR
	s.PublicSubnetName = key.SubnetName(cfg.CustomObject, suffixPublic)
	s.PublicSubnetMapPublicIPOnLaunch = false
	s.PrivateSubnetAZ = key.AvailabilityZone(cfg.CustomObject)
	s.PrivateSubnetCIDR = key.PrivateSubnetCIDR(cfg.CustomObject)
	s.PrivateSubnetName = key.SubnetName(cfg.CustomObject, suffixPrivate)
	s.PrivateSubnetMapPublicIPOnLaunch = false

	return nil
}
