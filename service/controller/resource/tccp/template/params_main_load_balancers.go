package template

type ParamsMainLoadBalancers struct {
	APIElbHealthCheckTarget          string
	APIElbName                       string
	APIInternalElbName               string
	APIElbPortsToOpen                []ParamsMainLoadBalancersPortPair
	APIElbScheme                     string
	APIInternalElbScheme             string
	APIElbSecurityGroupID            string
	EtcdElbHealthCheckTarget         string
	EtcdElbName                      string
	EtcdElbPortsToOpen               []ParamsMainLoadBalancersPortPair
	EtcdElbScheme                    string
	EtcdElbSecurityGroupID           string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	MasterInstanceResourceName       string
	PublicSubnets                    []string
	PrivateSubnets                   []string
}

type ParamsMainLoadBalancersPortPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}
