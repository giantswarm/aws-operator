package provider

import "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

type Interface interface {
	CurrentStatus() (v1alpha1.StatusCluster, error)
	CurrentVersion() (string, error)
	NextVersion() (string, error)
	UpdateVersion(nextVersion string) error
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
