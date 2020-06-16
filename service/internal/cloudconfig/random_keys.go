package cloudconfig

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
)

type RandomKeysConfig struct {
	Response map[string]randomkeys.Cluster
}

type RandomKeys struct {
	response map[string]randomkeys.Cluster
}

func NewRandomKeys(config RandomKeysConfig) (*RandomKeys, error) {
	r := &RandomKeys{
		response: config.Response,
	}

	return r, nil
}

func (r *RandomKeys) SearchCluster(id string) (randomkeys.Cluster, error) {
	c, ok := r.response[id]
	if ok {
		return c, nil
	}

	return randomkeys.Cluster{}, microerror.Mask(notFoundError)
}
