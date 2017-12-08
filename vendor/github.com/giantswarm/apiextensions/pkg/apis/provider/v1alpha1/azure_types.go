package v1alpha1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewAzureConfigCRD returns a new custom resource definition for AzureConfig.
// This might look something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: azureconfigs.provider.giantswarm.io
//     spec:
//       group: provider.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: AzureConfig
//         plural: azureconfigs
//         singular: azureconfig
//
func NewAzureConfigCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "azureconfigs.provider.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "provider.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "AzureConfig",
				Plural:   "azureconfigs",
				Singular: "azureconfig",
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AzureConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AzureConfigSpec `json:"spec"`
}

type AzureConfigSpec struct {
	Cluster Cluster              `json:"cluster" yaml:"cluster"`
	Azure   AzureConfigSpecAzure `json:"azure" yaml:"azure"`
}

type AzureConfigSpecAzure struct {
	DNSZones    AzureConfigSpecAzureDNSZones    `json:"dnsZones" yaml:"dnsZones"`
	HostCluster AzureConfigSpecAzureHostCluster `json:"hostCluster" yaml:"hostCluster"`
	// Location is the region for the resource group.
	Location string `json:"location" yaml:"location"`
	// StorageSKUName is the name of storage capability.
	// https://docs.microsoft.com/en-us/rest/api/storagerp/StorageAccounts/Create#definitions_skuname
	StorageSKUName string                             `json:"storageSKUName" yaml:"storageSKUName"`
	VirtualNetwork AzureConfigSpecAzureVirtualNetwork `json:"virtualNetwork" yaml:"virtualNetwork"`

	Masters []AzureConfigSpecAzureNode `json:"masters" yaml:"masters"`
	Workers []AzureConfigSpecAzureNode `json:"workers" yaml:"workers"`
}

// AzureConfigSpecAzureDNSZones contains the DNS Zones of the cluster.
type AzureConfigSpecAzureDNSZones struct {
	// API is the DNS Zone for the Kubernetes API.
	API AzureConfigSpecAzureDNSZonesDNSZone `json:"api" yaml:"api"`
	// Etcd is the DNS Zone for the etcd cluster.
	Etcd AzureConfigSpecAzureDNSZonesDNSZone `json:"etcd" yaml:"etcd"`
	// Ingress is the DNS Zone for the Ingress resource, used for customer traffic.
	Ingress AzureConfigSpecAzureDNSZonesDNSZone `json:"ingress" yaml:"ingress"`
}

// AzureConfigSpecAzureDNSZonesDNSZone points to a DNS Zone in Azure.
type AzureConfigSpecAzureDNSZonesDNSZone struct {
	// ResourceGroup is the resource group of the zone.
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
	// Name is the name of the zone.
	Name string `json:"name" yaml:"name"`
}

type AzureConfigSpecAzureHostCluster struct {
	// CIDR is the CIDR of the host cluster Virtual Network. This is going
	// to be used by the Guest Cluster to allow SSH traffic from that CIDR.
	CIDR string `json:"cidr" yaml:"cidr"`
	// ResourceGroup is the resource group name of the host cluster. It is
	// used to determine DNS hosted zone to put NS records in.
	ResourceGroup string `json:"resourceGroup" yaml:"resourceGroup"`
}

type AzureConfigSpecAzureVirtualNetwork struct {
	// CIDR is the CIDR for the Virtual Network.
	CIDR string `json:"cidr" yaml:"cidr"`
	// MasterSubnetCIDR is the CIDR for the master subnet,
	MasterSubnetCIDR string `json:"masterSubnetCIDR" yaml:"masterSubnetCIDR"`
	// WorkerSubnetCIDR is the CIDR for the worker subnet,
	WorkerSubnetCIDR string `json:"workerSubnetCIDR" yaml:"workerSubnetCIDR"`
}

type AzureConfigSpecAzureNode struct {
	// AdminUsername is the vm administrator username
	AdminUsername string `json:"adminUsername" yaml:"adminUsername"`
	//  AdminSSHKeyData is the vm administrator ssh public key
	AdminSSHKeyData string `json:"adminSSHKeyData" yaml:"adminSSHKeyData"`
	// DataDiskSizeGB is the vm data disk size in GB
	DataDiskSizeGB int `json:"dataDiskSizeGB" yaml:"dataDiskSizeGB"`
	// OSImage is the vm OS image object
	OSImage AzureConfigSpecAzureNodeOSImage `json:"osImage" yaml:"osImage"`
	// VMSize is the master vm size (e.g. Standard_A1)
	VMSize string `json:"vmSize" yaml:"vmSize"`
}

type AzureConfigSpecAzureNodeOSImage struct {
	// Publisher is the image publisher (e.g GiantSwarm)
	Publisher string `json:"publisher" yaml:"publisher"`
	// Offer is the image offered by the publisher (e.g. CoreOS)
	Offer string `json:"offer" yaml:"offer"`
	// SKU is the image SKU (e.g. Alpha)
	SKU string `json:"sku" yaml:"sku"`
	// Version is the image version (e.g. 1465.7.0)
	Version string `json:"version" yaml:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AzureConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AzureConfig `json:"items"`
}
