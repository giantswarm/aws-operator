package cloudconfig

type TemplateData struct {
	AWSCCMVersion         string
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
