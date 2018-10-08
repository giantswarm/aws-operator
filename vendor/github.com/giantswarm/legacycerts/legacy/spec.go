package legacy

type Searcher interface {
	SearchCerts(clusterID string) (AssetsBundle, error)
}
