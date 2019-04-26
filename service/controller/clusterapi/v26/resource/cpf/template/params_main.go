package template

// ParamsMain is the data structure for the Control Plane Finalizer template.
type ParamsMain struct {
	RecordSets  *ParamsMainRecordSets
	RouteTables *ParamsMainRouteTables
}
