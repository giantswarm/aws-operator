package randomkeytpr

type Searcher interface {
	SearchKeys(clusterID string) (map[Key][]byte, error)
	SearchKeysForKeytype(clusterID, keyType string) (map[Key][]byte, error)
}

type Spec struct {
	ClusterComponent string `json:"clusterComponent" yaml:"clusterComponent"`
	ClusterID        string `json:"clusterID" yaml:"clusterID"`
}
