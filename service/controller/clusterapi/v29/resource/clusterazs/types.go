package clusterazs

import "net"

// subnetPair is temporary type for mapping existing subnets from
// controllercontext to AZs.
type subnetPair struct {
	public  net.IPNet
	private net.IPNet
}

func (sp subnetPair) areEmpty() bool {
	return (sp.public.IP == nil && sp.public.Mask == nil) && (sp.private.IP == nil && sp.private.Mask == nil)
}
