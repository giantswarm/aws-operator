package bridgezone

import (
	"context"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/v13/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v13/credential"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

const (
	name = "bridgezonev13"
)

type Config struct {
	HostAWSConfig aws.Config
	HostRoute53   *route53.Route53
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger

	Route53Enabled bool
}

// Resource is bridgezone resource making sure we have fallback delegation in
// old DNS structure. This is only for the migration period. When we delete the
// "intermediate" zone this resource becomes noop and we do not need it
// anymore.
//
// Old structure looks like:
//
//	installation.eu-central-1.aws.gigantic.io
//	└── NS installation.k8s.eu-central-1.aws.gigantic.io
//	    ├── A api.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// New structure looks like:
//
//	installation.eu-central-1.aws.gigantic.io
//	└── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    ├── A api.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// For the migration period for new clusters we need also to add delegation to
// k8s.eu-central-1.aws.gigantic.io because of the AWS DNS caching issues.
//
//	installation.eu-central-1.aws.gigantic.io
//	├── NS k8s.installation.eu-central-1.aws.gigantic.io
//	│   ├── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	│   ├── A api.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	│   └── A ingress.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	└── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    ├── A api.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// NOTE: In the code below k8s.eu-central-1.aws.gigantic.io zone is called
// "intermediate" and new_cluster.k8s.eu-central-1.aws.gigantic.io is called
// "final". This resource ensures we have delegation from the intermediate zone
// to the final zone, but only if the intermediate zone exists.
//
// After we have guest clusters managed by this resource set (v13) and newer it
// means we can delete delegation to
// k8s.installation.eu-central-1.aws.gigantic.io from
// installation.eu-central-1.aws.gigantic.io zone. Then after a couple of days
// when delegation propagates and DNS caches are refreshed we can delete
// k8s.installation.eu-central-1.aws.gigantic.io zone.
type Resource struct {
	hostAWSConfig aws.Config
	hostRoute53   *route53.Route53
	k8sClient     kubernetes.Interface
	logger        micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(aws.Config{}, config.HostAWSConfig) {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostAWSConfig must not be empty", config)
	}
	if config.HostRoute53 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostRoute53 must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		hostAWSConfig: config.HostAWSConfig,
		hostRoute53:   config.HostRoute53,
		k8sClient:     config.K8sClient,
		logger:        config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return name
}

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled, skipping execution")
		return nil
	}

	baseDomain := key.BaseDomain(customObject)
	// TODO use key package and use it in tempates.
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(customObject) + ".k8s." + baseDomain

	guest, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "searching for intermediate zone in default guest account")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			// If the intermeidate zone is not found we are after
			// the migraiton period and this resource becomes noop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone does not exist, skipping the resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone found")
	}

	var finalZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "searching for final zone in guest account")

		finalZoneID, err = r.findHostedZoneID(ctx, defaultGuest, finalZone)
		if IsNotFound(err) {
			// The final zone is not yet created. Retry in the next
			// reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zone not found, skipping the resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "final zone found")
	}

	var finalZoneNS []string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone name servers")

		finalZoneNS, err = r.getNameServers(ctx, guest, finalZoneID, finalZone)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got name servers of final zone: %v")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring intermediate zone delegation")

		ttl := int64(900)
		var values []*route53.ResourceRecord
		for _, ns := range finalZoneNS {
			v := &route53.ResourceRecord{
				Value: &ns,
			}
			values = append(values, v)
		}

		upsert := route53.ChangeActionUpsert
		ns := route53.RRTypeNs

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &upsert,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name:            &finalZone,
							Type:            &ns,
							TTL:             &ttl,
							ResourceRecords: values,
						},
					},
				},
			},
			HostedZoneId: &intermediateZoneID,
		}
		_, err := defaultGuest.ChangeResourceRecordSetsWithContext(ctx, in)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured intermediate zone delegation")
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled, skipping execution")
		return nil
	}

	baseDomain := key.BaseDomain(customObject)
	// TODO use key package and use it in tempates.
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(customObject) + ".k8s." + baseDomain

	_, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "searching for intermediate zone in default guest account")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone does not exist, skipping the resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone found")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of intermediate zone delegation")

		delete := route53.ChangeActionDelete
		ns := route53.RRTypeNs

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &delete,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name: &finalZone,
							Type: &ns,
						},
					},
				},
			},
			HostedZoneId: &intermediateZoneID,
		}
		_, err := defaultGuest.ChangeResourceRecordSetsWithContext(ctx, in)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deletion of intermediate zone delegation ensured")
	}

	return nil
}

func (r *Resource) findHostedZoneID(ctx context.Context, client *route53.Route53, name string) (string, error) {
	var marker *string
	for {
		in := &route53.ListHostedZonesInput{
			Marker: marker,
		}

		out, err := client.ListHostedZones(in)
		if err != nil {
			return "", microerror.Mask(err)
		}

		for _, hz := range out.HostedZones {
			if hz.Name == nil || hz.Id == nil {
				continue
			}

			hzName := *hz.Name
			hzName = strings.TrimSuffix(hzName, ".")
			hzID := *hz.Id

			if hzName == name {
				return hzID, nil
			}
		}

		// If not all IDs are found, try to search next page.
		if out.IsTruncated == nil || !*out.IsTruncated {
			return "", microerror.Maskf(notFoundError, "HostedZone with name %q not found", name)
		}
		marker = out.Marker
	}
}

func (r *Resource) getNameServers(ctx context.Context, client *route53.Route53, zoneID, name string) ([]string, error) {
	one := "1"
	ns := route53.RRTypeNs
	in := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &zoneID,
		MaxItems:        &one,
		StartRecordName: &name,
		StartRecordType: &ns,
	}
	out, err := client.ListResourceRecordSetsWithContext(ctx, in)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(out.ResourceRecordSets) == 0 {
		return nil, microerror.Maskf(executionError, "NS recrod %q for HostedZone %q not found", name, zoneID)
	}
	if len(out.ResourceRecordSets) != 1 {
		return nil, microerror.Maskf(executionError, "expected single NS recrod %q for HostedZone %q, found %#v", name, zoneID, out.ResourceRecordSets)
	}

	rs := *out.ResourceRecordSets[0]

	if *rs.Name != name {
		return nil, microerror.Maskf(executionError, "expected NS recrod with name %q , found %q", name, *rs.Name)
	}

	var servers []string
	for _, r := range rs.ResourceRecords {
		servers = append(servers, *r.Value)
	}

	return servers, nil
}

func (r *Resource) route53Clients(ctx context.Context) (guest, defaultGuest *route53.Route53, err error) {
	// guest
	{
		controllerCtx, err := controllercontext.FromContext(ctx)
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}
		guest = controllerCtx.AWSClient.Route53
	}

	// defaultGuest
	{
		arn, err := credential.GetDefaultARN(r.k8sClient)
		if err != nil {
			return nil, nil, microerror.Mask(err)
		}

		c := r.hostAWSConfig
		c.RoleARN = arn
		defaultGuest = aws.NewClients(c).Route53
	}

	return guest, defaultGuest, nil
}
