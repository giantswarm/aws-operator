package cloudconfig

type TemplateData struct {
	AWSRegion            string
	IsChinaRegion        bool
	MasterENIAddress     string
	MasterENIGateway     string
	MasterENIName        string
	MasterENISubnetSize  int
	MasterEtcdVolumeName string
	MasterID             int
	RegistryDomain       string
}
