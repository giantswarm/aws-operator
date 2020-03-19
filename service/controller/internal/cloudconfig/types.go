package cloudconfig

import g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

type TemplateData struct {
	AWSRegion           string
	AWSConfigSpec       g8sv1alpha1.AWSConfigSpec
	IsChinaRegion       bool
	MasterENIAddresses  []string
	MasterENIGateways   []string
	MasterENISubnetSize string
	MasterID            int
	RegistryDomain      string
}
