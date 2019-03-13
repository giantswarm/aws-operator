package template

type ParamsMainSubnets []ParamsMainSubnetsSubnet

type ParamsMainSubnetsSubnet struct {
	AvailabilityZone string
	CIDR             string
	Name             string
}
