package adapter

// template related to this adapter: service/templates/cloudformation/guest/recordsets.yaml

type recordSetsAdapter struct {
	APIELBHostedZones         string
	APIELBDomain              string
	EtcdELBHostedZones        string
	EtcdELBDomain             string
	IngressELBDNS             string
	IngressELBHostedZones     string
	IngressELBAliasHostedZone string
	IngressELBDomain          string
	IngressWildcardELBDomain  string
}

func (r *recordSetsAdapter) getRecordSets(cfg Config) error {
	r.APIELBHostedZones = cfg.CustomObject.Spec.AWS.API.HostedZones
	r.APIELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain
	r.EtcdELBHostedZones = cfg.CustomObject.Spec.AWS.Etcd.HostedZones
	r.EtcdELBDomain = cfg.CustomObject.Spec.Cluster.Etcd.Domain
	r.IngressELBHostedZones = cfg.CustomObject.Spec.AWS.Ingress.HostedZones
	r.IngressELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain
	r.IngressWildcardELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.WildcardDomain

	return nil
}
