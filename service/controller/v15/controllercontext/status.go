package controllercontext

type Status struct {
	// Cluster carries an information between cluster controller resource.
	Cluster Cluster
	// Drainer carries an information between drainer controller resource.
	Drainer Drainer
}

type Cluster struct {
	// HostedZones is filled by the hostedzone resource. This information
	// is used when creating CloudFormation templates.
	HostedZones ClusterHostedZones
}

type ClusterHostedZones struct {
	API     ClusterHostedZonesZone
	Etcd    ClusterHostedZonesZone
	Ingress ClusterHostedZonesZone
}

type ClusterHostedZonesZone struct {
	ID string
}

type Drainer struct {
	// WorkerASGName is filled by the workerasgname resource.
	WorkerASGName string
}
