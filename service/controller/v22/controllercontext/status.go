package controllercontext

type Status struct {
	// Cluster carries information between cluster controller resource.
	Cluster Cluster
	// Drainer carries information between drainer controller resource.
	Drainer Drainer
}

type Cluster struct {
	AWSAccount    ClusterAWSAccount
	ASG           ClusterASG
	EncryptionKey string
}

type ClusterAWSAccount struct {
	ID string
}

type ClusterASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
}

type Drainer struct {
	// WorkerASGName is filled by the workerasgname resource.
	WorkerASGName string
}
