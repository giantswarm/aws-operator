package annotation

const (
	Docs                    = "giantswarm.io/docs"
	InstanceID              = "aws-operator.giantswarm.io/instance"
	MachineDeploymentSubnet = "machine-deployment.giantswarm.io/subnet"
	// AWSMetadataV2 configures token usage for your AWS EC2 instance metadata requests.
	// If the value is 'optional', you can choose to retrieve instance metadata with or without a signed token
	// header on your request. If you retrieve the IAM role credentials without a token, the version 1.0 role
	// credentials are returned. If you retrieve the IAM role credentials using a valid signed token, the version
	// 2.0 role credentials are returned.
	// If the state is 'required', you must send a signed token header with any instance metadata retrieval
	// requests. In this state, retrieving the IAM role credentials always returns the version 2.0 credentials; the
	// version 1.0 credentials are not available.
	// Default: 'optional'
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ec2-launchtemplate-launchtemplatedata-metadataoptions.html#cfn-ec2-launchtemplate-launchtemplatedata-metadataoptions-httptokens
	AWSMetadata = "alpha.aws.giantswarm.io/metadata-v2"
)
