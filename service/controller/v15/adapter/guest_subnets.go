package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/v15/key"
)

type GuestSubnetsAdapter struct {
	PublicSubnetAZ                   string
	PublicSubnetCIDR                 string
	PublicSubnetName                 string
	PublicSubnetMapPublicIPOnLaunch  bool
	PrivateSubnetAZ                  string
	PrivateSubnetCIDR                string
	PrivateSubnetName                string
	PrivateSubnetMapPublicIPOnLaunch bool
}

func (s *GuestSubnetsAdapter) Adapt(cfg Config) error {
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
