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
	"github.com/giantswarm/aws-operator/service/controller/v14/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v14/credential"
	"github.com/giantswarm/aws-operator/service/controller/v14/key"
)

const (
	name = "bridgezonev14"
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
//	installation.eu-central-1.aws.gigantic.io (host account)
//	└── NS k8s.installation.eu-central-1.aws.gigantic.io (default guest account)
//	    ├── A api.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// New structure looks like:
//
//	installation.eu-central-1.aws.gigantic.io (host account)
//	└── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io (byoc guest account)
//	    ├── A api.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// For the migration period for new clusters we need also to add delegation to
// k8s.installation.eu-central-1.aws.gigantic.io because of the AWS DNS caching issues.
//
//	installation.eu-central-1.aws.gigantic.io (host account)
//	├── NS k8s.installation.eu-central-1.aws.gigantic.io (default guest account)
//	│   ├── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io (byoc guest account)
//	│   ├── A api.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	│   └── A ingress.old_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	└── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io (byoc guest account)
//	    ├── A api.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// NOTE: In the code below k8s.installation.eu-central-1.aws.gigantic.io zone is called
// "intermediate" and new_cluster.k8s.installation.eu-central-1.aws.gigantic.io zone is
// called "final". This resource *only* ensures we have delegation from the
// intermediate zone to the final zone, but only if the intermediate zone
// exists.
//
// After we have guest clusters managed by this resource set (v14) and newer
// the DNS layout should look like:
//
//	installation.eu-central-1.aws.gigantic.io (host account)
//	├── NS k8s.installation.eu-central-1.aws.gigantic.io (default guest account)
//	│   └── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io (byoc guest account)
//	└── NS new_cluster.k8s.installation.eu-central-1.aws.gigantic.io (byoc guest account)
//	    ├── A api.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//	    └── A ingress.new_cluster.k8s.installation.eu-central-1.aws.gigantic.io
//
// At this point we should be fine with removing
// k8s.installation.eu-central-1.aws.gigantic.io NS record from
// installation.eu-central-1.aws.gigantic.io zone. Then after a couple of days
// when delegation propagates and DNS caches are refreshed we can delete
// k8s.installation.eu-central-1.aws.gigantic.io zone from the default guest
// account.
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
		return nil
	}

	baseDomain := key.BaseDomain(customObject)
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(customObject) + ".k8s." + baseDomain

	guest, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting intermediate zone ID")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			// If the intermeidate zone is not found we are after
			// the migraiton period and this resource becomes noop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got intermediate zone ID")
	}

	var finalZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone ID")

		finalZoneID, err = r.findHostedZoneID(ctx, defaultGuest, finalZone)
		if IsNotFound(err) {
			// The final zone is not yet created. Retry in the next
			// reconciliation loop.
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zone not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone ID")
	}

	var finalZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone name servers")

		nameServers, _, err := r.getNameServersAndTTL(ctx, guest, finalZoneID, finalZone)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalZoneRecords = append(finalZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone name servers")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring final zone delegation from intermediate zone")

		upsert := route53.ChangeActionUpsert
		ns := route53.RRTypeNs
		ttl := int64(900)

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &upsert,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name:            &finalZone,
							Type:            &ns,
							TTL:             &ttl,
							ResourceRecords: finalZoneRecords,
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

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured final zone delegation from intermediate zone")
	}

	return nil
}

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !r.route53Enabled {
		r.logger.LogCtx(ctx, "level", "debug", "message", "route53 disabled")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
		return nil
	}

	baseDomain := key.BaseDomain(customObject)
	intermediateZone := "k8s." + baseDomain
	finalZone := key.ClusterID(customObject) + ".k8s." + baseDomain

	_, defaultGuest, err := r.route53Clients(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var intermediateZoneID string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting intermediate zone ID")

		intermediateZoneID, err = r.findHostedZoneID(ctx, defaultGuest, intermediateZone)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "intermediate zone not found")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got intermediate zone ID")
	}

	var finalZoneTTL int64
	var finalZoneRecords []*route53.ResourceRecord
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "getting final zone delegation name servers and TTL from intermediate zone")

		nameServers, ttl, err := r.getNameServersAndTTL(ctx, defaultGuest, intermediateZoneID, finalZone)
		if IsNotFound(err) {
			// Delegation may be already deleted. It must be handled.
			r.logger.LogCtx(ctx, "level", "debug", "message", "final zone delegation not found in intermediate zone")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		finalZoneTTL = ttl

		for _, ns := range nameServers {
			copy := ns
			v := &route53.ResourceRecord{
				Value: &copy,
			}
			finalZoneRecords = append(finalZoneRecords, v)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "got final zone delegation name servers and TTL from intermediate zone")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of final zone delegation from intermediate zone")

		delete := route53.ChangeActionDelete
		ns := route53.RRTypeNs

		in := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: &delete,
						ResourceRecordSet: &route53.ResourceRecordSet{
							Name:            &finalZone,
							Type:            &ns,
							TTL:             &finalZoneTTL,
							ResourceRecords: finalZoneRecords,
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

		r.logger.LogCtx(ctx, "level", "debug", "message", "ensured deletion of final zone delegation from intermediate zone")
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

func (r *Resource) getNameServersAndTTL(ctx context.Context, client *route53.Route53, zoneID, name string) (nameServers []string, ttl int64, err error) {
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
		return nil, 0, microerror.Mask(err)
	}

	if len(out.ResourceRecordSets) == 0 {
		return nil, 0, microerror.Maskf(notFoundError, "NS recrod %q for HostedZone %q not found", name, zoneID)
	}
	if len(out.ResourceRecordSets) != 1 {
		return nil, 0, microerror.Maskf(executionError, "expected single NS recrod %q for HostedZone %q, found %#v", name, zoneID, out.ResourceRecordSets)
	}

	rs := *out.ResourceRecordSets[0]

	if strings.TrimSuffix(*rs.Name, ".") != name {
		return nil, 0, microerror.Maskf(executionError, "expected NS recrod with name %q , found %q", name, *rs.Name)
	}

	var servers []string
	for _, r := range rs.ResourceRecords {
		servers = append(servers, *r.Value)
	}

	return servers, *rs.TTL, nil
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
