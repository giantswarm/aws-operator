package template

type ParamsLoadBalancers struct {
	APIElbHealthCheckTarget          string
	APIElbName                       string
	APIInternalElbName               string
	APIElbPortsToOpen                []GuestLoadBalancersAdapterPortPair
	APIElbScheme                     string
	APIInternalElbScheme             string
	APIElbSecurityGroupID            string
	EtcdElbHealthCheckTarget         string
	EtcdElbName                      string
	EtcdElbPortsToOpen               []GuestLoadBalancersAdapterPortPair
	EtcdElbScheme                    string
	EtcdElbSecurityGroupID           string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	IngressElbHealthCheckTarget      string
	IngressElbName                   string
	IngressElbPortsToOpen            []GuestLoadBalancersAdapterPortPair
	IngressElbScheme                 string
	MasterInstanceResourceName       string
	PublicSubnets                    []string
	PrivateSubnets                   []string
}

type ParamsLoadBalancersPortPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}
