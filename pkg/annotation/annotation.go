package annotation

const (
	Docs                    = "giantswarm.io/docs"
	InstanceID              = "aws-operator.giantswarm.io/instance"
	MachineDeploymentSubnet = "machine-deployment.giantswarm.io/subnet"
	// Annotation used to enable feature to terminate unhealthy nodes on a cluster CR.
	NodeTerminateUnhealthy = "alpha.node.giantswarm.io/terminate-unhealthy"
	AWSMetadata             = "alpha.giantswarm.io/aws-metadata-v2"
)
