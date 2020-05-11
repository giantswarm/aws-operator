package cloudconfig

type TemplateData struct {
	AWSRegion            string
	IsChinaRegion        bool
	MasterENIName        string
	MasterEtcdVolumeName string
	MasterID             int
	RegistryDomain       string
}
