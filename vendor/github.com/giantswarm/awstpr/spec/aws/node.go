package aws

type Node struct {
	// ImageID of EC2 instance.
	ImageID string `json:"imageID"`
	// InstanceType of EC2 instance.
	InstanceType string `json:"instanceType"`
}
