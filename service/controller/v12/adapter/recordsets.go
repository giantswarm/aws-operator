package adapter

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v12/templates/cloudformation/guest/recordsets.go
//

type recordSetsAdapter struct {
	APIELBHostedZones string
	APIELBDomain      string
	Route53Enabled    bool
}

func (r *recordSetsAdapter) getRecordSets(cfg Config) error {
	r.APIELBHostedZones = cfg.CustomObject.Spec.AWS.API.HostedZones
	r.APIELBDomain = cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain
	r.Route53Enabled = cfg.Route53Enabled

	return nil
}
