package template

type ParamsMain struct {
	EnableAWSCNI    bool
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
