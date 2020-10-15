package template

type ParamsMain struct {
	InternetGateway *ParamsMainInternetGateway
	LoadBalancers   *ParamsMainLoadBalancers
	NATGateway      *ParamsMainNATGateway
	Outputs         *ParamsMainOutputs
	RecordSets      *ParamsMainRecordSets
	RouteTables     *ParamsMainRouteTables
	SecurityGroups  *ParamsMainSecurityGroups
	Subnets         *ParamsMainSubnets
	VPC             *ParamsMainVPC
}
