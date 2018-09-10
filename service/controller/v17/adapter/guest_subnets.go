package adapter

import (
	"net"

	"github.com/giantswarm/aws-operator/service/controller/v17/key"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
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
	publicSubnet, privateSubnet, err := allocatePublicAndPrivateSubnets(cfg)
	if err != nil {
		return microerror.Mask(err)
	}

	s.PublicSubnetAZ = key.AvailabilityZone(cfg.CustomObject)
	s.PublicSubnetCIDR = publicSubnet.String()
	s.PublicSubnetName = key.SubnetName(cfg.CustomObject, suffixPublic)
	s.PublicSubnetMapPublicIPOnLaunch = false
	s.PrivateSubnetAZ = key.AvailabilityZone(cfg.CustomObject)
	s.PrivateSubnetCIDR = privateSubnet.String()
	s.PrivateSubnetName = key.SubnetName(cfg.CustomObject, suffixPrivate)
	s.PrivateSubnetMapPublicIPOnLaunch = false

	return nil
}

func allocatePublicAndPrivateSubnets(cfg Config) (net.IPNet, net.IPNet, error) {
	_, subnet, err := net.ParseCIDR(key.ClusterNetworkCIDR(cfg.CustomObject))
	if err != nil {
		return net.IPNet{}, net.IPNet{}, microerror.Mask(err)
	}

	privateSubnetMask := net.CIDRMask(cfg.PrivateSubnetMaskBits, 32)
	publicSubnetMask := net.CIDRMask(cfg.PublicSubnetMaskBits, 32)

	var reservedSubnets []net.IPNet
	privateSubnet, err := ipam.Free(*subnet, privateSubnetMask, reservedSubnets)
	if err != nil {
		return net.IPNet{}, net.IPNet{}, microerror.Mask(err)
	}

	reservedSubnets = append(reservedSubnets, privateSubnet)

	publicSubnet, err := ipam.Free(*subnet, publicSubnetMask, reservedSubnets)
	if err != nil {
		return net.IPNet{}, net.IPNet{}, microerror.Mask(err)
	}

	return publicSubnet, privateSubnet, nil
}
