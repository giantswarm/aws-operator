package template

type ParamsMainENI struct {
	List []ParamsMainENIItem
}

type ParamsMainENIItem struct {
	Name            string
	Resource        string
	SecurityGroupID string
	SubnetID        string
}
