package aws

type Node struct {
	// ImageID of EC2 instance.
	ImageID string `json:"imageID" yaml:"imageID"`
	// InstanceType of EC2 instance.
	InstanceType string `json:"instanceType" yaml:"instanceType"`
}
