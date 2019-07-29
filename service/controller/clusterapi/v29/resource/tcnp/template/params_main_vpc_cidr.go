package template

type ParamsMainVPCCIDR struct {
	TCCP ParamsMainVPCCIDRTCCP
	TCNP ParamsMainVPCCIDRTCNP
}

type ParamsMainVPCCIDRTCCP struct {
	VPC ParamsMainVPCCIDRTCCPVPC
}

type ParamsMainVPCCIDRTCCPVPC struct {
	ID string
}

type ParamsMainVPCCIDRTCNP struct {
	CIDR string
}
