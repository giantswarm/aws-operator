package certificatetpr

import "github.com/giantswarm/versionbundle"

type Searcher interface {
	SearchCerts(clusterID string) (AssetsBundle, error)
}

type Spec struct {
	AllowBareDomains bool                 `json:"allowBareDomains" yaml:"allowBareDomains"`
	AltNames         []string             `json:"altNames" yaml:"altNames"`
	ClusterComponent string               `json:"clusterComponent" yaml:"clusterComponent"`
	ClusterID        string               `json:"clusterID" yaml:"clusterID"`
	CommonName       string               `json:"commonName" yaml:"commonName"`
	IPSANs           []string             `json:"ipSans" yaml:"ipSans"`
	Organizations    []string             `json:"organizations" yaml:"organizations"`
	TTL              string               `json:"ttl" yaml:"ttl"`
	VersionBundle    versionbundle.Bundle `json:"version_bundle" yaml:"version_bundle"`
}
