package adapter

// CPFConfig represents the config for the adapter collection for the Control
// Plane Finalizer management.
type CPFConfig struct {
	BaseDomain                 string
	ClusterID                  string
	GuestHostedZoneNameServers string
	PrivateRoutes              []CPFRouteTablesRoute
	PublicRoutes               []CPFRouteTablesRoute
	Route53Enabled             bool
}

// CPF is the adapter collection for the Control Plane Finalizer management.
type CPF struct {
	RecordSets  *CPFRecordSets
	RouteTables *CPFRouteTables
}

func NewCPF(config CPFConfig) (*CPF, error) {
	var recordSets *CPFRecordSets
	{
		recordSets = &CPFRecordSets{
			BaseDomain:                 config.BaseDomain,
			ClusterID:                  config.ClusterID,
			GuestHostedZoneNameServers: config.GuestHostedZoneNameServers,
			Route53Enabled:             config.Route53Enabled,
		}
	}

	var routeTables *CPFRouteTables
	{
		routeTables = &CPFRouteTables{
			PrivateRoutes: config.PrivateRoutes,
			PublicRoutes:  config.PublicRoutes,
		}
	}

	cpf := &CPF{
		RecordSets:  recordSets,
		RouteTables: routeTables,
	}

	return cpf, nil
}
