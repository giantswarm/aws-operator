package delete

import (
	"github.com/giantswarm/aws-operator/service/common"
)

type Config struct {
	common.Config
}

func New(config Config) (*Service, error) {
	return nil, nil
}

type Service struct {
	common.Service
}
