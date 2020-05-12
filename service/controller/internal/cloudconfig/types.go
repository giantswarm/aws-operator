package cloudconfig

type TemplateData struct {
	AWSRegion            string
	ExternalSNAT         bool
	IsChinaRegion        bool
	MasterENIAddress     string
	MasterENIGateway     string
	MasterENIName        string
	MasterENISubnetSize  int
	MasterEtcdVolumeName string
	MasterID             int
	RegistryDomain       string
}
