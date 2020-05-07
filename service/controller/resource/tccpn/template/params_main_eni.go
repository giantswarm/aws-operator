package template

type ParamsMainENI struct {
	List []ParamsMainENIItem
}

type ParamsMainENIItem struct {
	IpAddress       string
	Name            string
	Resource        string
	SecurityGroupID string
	SubnetID        string
}
