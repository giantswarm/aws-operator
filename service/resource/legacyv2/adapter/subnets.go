package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

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

func (s *subnetsAdapter) getSubnets(customObject v1alpha1.AWSConfig, clients Clients) error {
	s.PublicSubnetAZ = keyv2.AvailabilityZone(customObject)
	s.PublicSubnetCIDR = customObject.Spec.AWS.VPC.PublicSubnetCIDR
	s.PublicSubnetName = keyv2.SubnetName(customObject, suffixPublic)
	s.PublicSubnetMapPublicIPOnLaunch = false
	s.PrivateSubnetAZ = keyv2.AvailabilityZone(customObject)
	s.PrivateSubnetCIDR = customObject.Spec.AWS.VPC.PrivateSubnetCIDR
	s.PrivateSubnetName = keyv2.SubnetName(customObject, suffixPrivate)
	s.PrivateSubnetMapPublicIPOnLaunch = false

	return nil
}
