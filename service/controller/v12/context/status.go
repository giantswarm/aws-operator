package context

type Status struct {
	// HostedZones information is filled by the hostedzone resouce. This
	// information is used when creating CloudFormation templates.
	HostedZones HostedZones
}

type HostedZones struct {
	API     HostedZonesZone
	Etcd    HostedZonesZone
	Ingress HostedZonesZone
}

type HostedZonesZone struct {
	ID string
}
