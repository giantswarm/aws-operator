package controllercontext

import "net"

type ContextTenantClusterAvailabilityZone struct {
	Name          string
	PrivateSubnet net.IPNet
	PublicSubnet  net.IPNet
}
