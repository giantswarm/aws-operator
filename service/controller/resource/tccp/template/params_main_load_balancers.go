package template

type ParamsMainLoadBalancers struct {
	APIElbHealthCheckTarget    string
	APIElbName                 string
	APIInternalElbName         string
	APIElbPortsToOpen          []ParamsMainLoadBalancersPortPair
	APIElbSecurityGroupID      string
	EtcdElbHealthCheckTarget   string
	EtcdElbName                string
	EtcdElbPortsToOpen         []ParamsMainLoadBalancersPortPair
	EtcdElbSecurityGroupID     string
	MasterInstanceResourceName string
	PublicSubnets              []string
	PrivateSubnets             []string
}

type ParamsMainLoadBalancersPortPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}
