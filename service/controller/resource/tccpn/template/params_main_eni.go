package template

type ParamsMainENI struct {
	ENIs []ParamsMainENISpec
}

type ParamsMainENISpec struct {
	IpAddress       string
	Name            string
	SecurityGroupID string
	SubnetID        string
	ResourceName    string
}
