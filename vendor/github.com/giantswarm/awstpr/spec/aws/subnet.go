package aws

type Subnet struct {
	// Name of the subnet.
	Name string `json:"name" yaml:"name"`
	// RouteTableName of the route table associated to the subnet.
	RouteTableName string `json:"routeTableName" yaml:"routeTableName"`
}
