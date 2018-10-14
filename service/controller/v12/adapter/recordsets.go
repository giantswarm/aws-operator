package adapter

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v12/templates/cloudformation/guest/recordsets.go
//

type recordSetsAdapter struct {
	APIELBDomain             string
	EtcdELBDomain            string
	IngressELBDNS            string
	IngressELBDomain         string
	IngressWildcardELBDomain string
	Route53Enabled           bool
}

func (r *recordSetsAdapter) getRecordSets(cfg Config) error {
	r.APIELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain
	r.EtcdELBDomain = cfg.CustomObject.Spec.Cluster.Etcd.Domain
	r.IngressELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain
	r.IngressWildcardELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.WildcardDomain
	r.Route53Enabled = cfg.Route53Enabled

	return nil
}
