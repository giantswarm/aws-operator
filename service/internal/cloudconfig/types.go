package cloudconfig

type TemplateData struct {
	AWSRegion            string
	BaseDomain           string
	ExternalSNAT         bool
	IsChinaRegion        bool
	MasterENIName        string
	MasterEtcdVolumeName string
	MasterID             int
	RegistryDomain       string
}
