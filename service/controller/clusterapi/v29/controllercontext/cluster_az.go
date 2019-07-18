package controllercontext

import "net"

type ContextTenantClusterAvailabilityZone struct {
	ID            string
	Name          string
	PrivateSubnet net.IPNet
	PublicSubnet  net.IPNet
}
