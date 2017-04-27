package aws

const (
	// EC2 instance tag keys.
	tagKeyName    string = "Name"
	tagKeyCluster string = "Cluster"
	// Subnet keys
	subnetAvailabilityZone string = "availabilityZone"
	subnetCidrBlock        string = "cidrBlock"
	subnetDescription      string = "description"
	subnetGroupName        string = "group-name"
	subnetVpcID            string = "vpc-id"
	// Security group keys
	securityGroupIPPermissionCIDR     string = "ip-permission.cidr"
	securityGroupIPPermissionFromPort string = "ip-permission.from-port"
	securityGroupIPPermissionToPort   string = "ip-permission.to-port"
)
