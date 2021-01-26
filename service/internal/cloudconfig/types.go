package cloudconfig

type TemplateData struct {
	AWSCNIMinimumIPTarget string
	AWSCNIVersion         string
	AWSCNIWarmIPTarget    string
	AWSRegion             string
	BaseDomain            string
	ExternalSNAT          bool
	IsChinaRegion         bool
	MasterENIName         string
	MasterEtcdVolumeName  string
	MasterID              int
	RegistryDomain        string
}
