package adapter

import (
	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/subnets.yaml

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
	s.PublicSubnetAZ = keyv2.AvailabilityZone(cfg.CustomObject)
	s.PublicSubnetCIDR = cfg.CustomObject.Spec.AWS.VPC.PublicSubnetCIDR
	s.PublicSubnetName = keyv2.SubnetName(cfg.CustomObject, suffixPublic)
	s.PublicSubnetMapPublicIPOnLaunch = false
	s.PrivateSubnetAZ = keyv2.AvailabilityZone(cfg.CustomObject)
	s.PrivateSubnetCIDR = cfg.CustomObject.Spec.AWS.VPC.PrivateSubnetCIDR
	s.PrivateSubnetName = keyv2.SubnetName(cfg.CustomObject, suffixPrivate)
	s.PrivateSubnetMapPublicIPOnLaunch = false

	return nil
}
