package template

// ParamsMain is the data structure for the Tenant Cluster Control Plane
// Finalizer template.
type ParamsMain struct {
	RecordSets  *ParamsMainRecordSets
	RouteTables *ParamsMainRouteTables
}
