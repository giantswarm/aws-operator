package v1alpha2

// +k8s:deepcopy-gen=false

type NodePoolCRsConfig struct {
	AvailabilityZones                   []string
	AWSInstanceType                     string
	ClusterID                           string
	MachineDeploymentID                 string
	Description                         string
	NodesMax                            int
	NodesMin                            int
	OnDemandBaseCapacity                int
	OnDemandPercentageAboveBaseCapacity int
	Owner                               string
	ReleaseComponents                   map[string]string
	ReleaseVersion                      string
	UseAlikeInstanceTypes               bool
}

// +k8s:deepcopy-gen=false

