package aws

const (
	// EC2 instance tag keys.
	tagKeyName    string = "Name"
	tagKeyCluster string = "KubernetesCluster"
	// Subnet keys
	subnetAvailabilityZone string = "availabilityZone"
	subnetCidrBlock        string = "cidrBlock"
	subnetDescription      string = "description"
	subnetGroupName        string = "group-name"
	subnetVpcID            string = "vpc-id"
	// Security Group IP Permission keys
	ipPermissionCIDR     string = "ip-permission.cidr"
	ipPermissionFromPort string = "ip-permission.from-port"
	ipPermissionGroupID  string = "ip-permission.group-id"
	ipPermissionProtocol string = "ip-permission.protocol"
	ipPermissionToPort   string = "ip-permission.to-port"
)
