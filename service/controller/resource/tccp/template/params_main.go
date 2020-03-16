package template

type ParamsMain struct {
	IAMPolicies     *ParamsMainIAMPolicies
	InternetGateway *ParamsMainInternetGateway
	Instance        *ParamsMainInstance
	LoadBalancers   *ParamsMainLoadBalancers
	NATGateway      *ParamsMainNATGateway
	Outputs         *ParamsMainOutputs
	RecordSets      *ParamsMainRecordSets
	RouteTables     *ParamsMainRouteTables
	SecurityGroups  *ParamsMainSecurityGroups
	Subnets         *ParamsMainSubnets
	VPC             *ParamsMainVPC
}
