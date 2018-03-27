package randomkeystest

import (
	"github.com/giantswarm/randomkeys"
)

type Searcher struct {
}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s *Searcher) SearchCluster(clusterID string) (randomkeys.Cluster, error) {
	return randomkeys.Cluster{}, nil
}
