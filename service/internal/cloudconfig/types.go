package cloudconfig

type TemplateData struct {
	AWSCNIAdditionalTags  string
	AWSCNIMinimumIPTarget string
	AWSCNIPrefix          bool
	AWSCNIWarmIPTarget    string
	AWSCNIVersion         string
	AWSRegion             string
	BaseDomain            string
	ExternalSNAT          bool
	IsChinaRegion         bool
	MasterENIName         string
	MasterEtcdVolumeName  string
	MasterID              int
	RegistryDomain        string
}
