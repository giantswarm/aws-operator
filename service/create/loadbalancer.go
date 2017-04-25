package create

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"
	"github.com/juju/errgo"
)

type LoadBalancerInput struct {
	Clients         awsutil.Clients
	Cluster         awstpr.CustomObject
	InstanceIDs     []string
	SecurityGroupID string
	SubnetID        string
}

func (s *Service) createLoadBalancer(input LoadBalancerInput) error {
	lb := &awsresources.ELB{
		Name:          input.Cluster.Spec.Cluster.Cluster.ID,
		SecurityGroup: input.SecurityGroupID,
		SubnetID:      input.SubnetID,
		PortsToOpen: []int{
			input.Cluster.Spec.Cluster.Kubernetes.API.SecurePort,
			input.Cluster.Spec.Cluster.Etcd.Port,
		},
		Client: input.Clients.ELB,
	}

	lbCreated, err := lb.CreateIfNotExists()
	if err != nil {
		return microerror.MaskAny(fmt.Errorf("could not create ELB: %s", errgo.Details(err)))
	}

	if lbCreated {
		s.logger.Log("debug", fmt.Sprintf("created ELB '%s'", lb.Name))
	} else {
		s.logger.Log("debug", fmt.Sprintf("ELB '%s' already exists, reusing", lb.Name))
	}

	// create DNS records for LB
	hzName, err := hostedZoneName(input.Cluster)
	if err != nil {
		return microerror.MaskAny(fmt.Errorf("could not generate hosted zone name: %s", err))
	}

	hz := awsresources.HostedZone{
		Name:    hzName,
		Comment: hostedZoneComment(input.Cluster),
		Client:  input.Clients.Route53,
	}

	hzCreated, err := hz.CreateIfNotExists()
	if err != nil {
		return microerror.MaskAny(fmt.Errorf("error creating hosted zone '%s'", errgo.Details(err)))
	}

	if hzCreated {
		s.logger.Log("debug", fmt.Sprintf("created hosted zone '%s'", hz.Name))
	} else {
		s.logger.Log("debug", fmt.Sprintf("hosted zone '%s' already exists, reusing", hz.Name))
	}

	apiRecordSet := &awsresources.RecordSet{
		Client:       input.Clients.Route53,
		Resource:     lb,
		Domain:       input.Cluster.Spec.Cluster.Kubernetes.API.Domain,
		HostedZoneID: hz.ID(),
	}

	if err := apiRecordSet.CreateOrFail(); err != nil {
		return microerror.MaskAny(fmt.Errorf("error registering DNS '%s'", errgo.Details(err)))
	}

	s.logger.Log("debug", "created or reused DNS record for api")

	etcdRecordSet := &awsresources.RecordSet{
		Client:       input.Clients.Route53,
		Resource:     lb,
		Domain:       input.Cluster.Spec.Cluster.Etcd.Domain,
		HostedZoneID: hz.ID(),
	}

	if err := etcdRecordSet.CreateOrFail(); err != nil {
		return microerror.MaskAny(fmt.Errorf("error registering DNS '%s'", errgo.Details(err)))
	}

	s.logger.Log("debug", "created or reused DNS record for etcd")

	s.logger.Log("debug", "waiting for masters to be ready...")

	var awsFlavouredInstanceIDs []*string
	for _, instanceID := range input.InstanceIDs {
		awsFlavouredInstanceIDs = append(awsFlavouredInstanceIDs, aws.String(instanceID))
	}

	if err := input.Clients.EC2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: awsFlavouredInstanceIDs,
	}); err != nil {
		return microerror.MaskAny(fmt.Errorf("masters took too long to get running, aborting: %v", err))
	}

	if err := lb.RegisterInstances(input.InstanceIDs); err != nil {
		return microerror.MaskAny(fmt.Errorf("could not register instances with LB: %s", errgo.Details(err)))
	}

	s.logger.Log("debug", fmt.Sprintf("instances registered with ELB"))

	return nil
}

func (s *Service) deleteLoadBalancer(input LoadBalancerInput) error {
	// Delete ELB
	lb, err := awsresources.NewExistingELB(input.Cluster.Spec.Cluster.Cluster.ID, input.Clients.ELB)
	if err != nil {
		return microerror.MaskAny(err)
	}

	if err := lb.Delete(); err != nil {
		return microerror.MaskAny(err)
	}
	s.logger.Log("debug", "deleted ELB")

	hzName, err := hostedZoneName(input.Cluster)
	if err != nil {
		return microerror.MaskAny(err)
	}

	hz, err := awsresources.NewExistingHostedZone(hzName, input.Clients.Route53)
	if err != nil {
		underlying := errgo.Cause(err)
		switch underlying.(type) {
		case awsresources.DomainNamedResourceNotFoundError:
			s.logger.Log("debug", "could not find existing hosted zone, continuing")
		default:
			return microerror.MaskAny(err)
		}
	}

	// Delete DNS record, if the Hosted Zone was found
	if hz != nil {
		apiRecordSet := &awsresources.RecordSet{
			Client:       input.Clients.Route53,
			Resource:     lb,
			Domain:       input.Cluster.Spec.Cluster.Kubernetes.API.Domain,
			HostedZoneID: hz.ID(),
		}

		if err := apiRecordSet.Delete(); err != nil {
			return microerror.MaskAny(err)
		}

		s.logger.Log("debug", "deleted DNS entry for api")

		etcdRecordSet := &awsresources.RecordSet{
			Client:       input.Clients.Route53,
			Resource:     lb,
			Domain:       input.Cluster.Spec.Cluster.Etcd.Domain,
			HostedZoneID: hz.ID(),
		}

		if err := etcdRecordSet.Delete(); err != nil {
			return microerror.MaskAny(err)
		}

		s.logger.Log("debug", "deleted DNS entry for etcd")
	}

	return nil
}

// hostedZoneName removes the last subdomain from the API domain
// e.g.  foobar.aws.giantswarm.io -> aws.giantswarm.io
func hostedZoneName(cluster awstpr.CustomObject) (string, error) {
	tmp := strings.SplitN(cluster.Spec.Cluster.Kubernetes.API.Domain, ".", 2)

	if len(tmp) == 0 {
		return "", microerror.MaskAny(malformedDNSNameError)
	}

	return strings.Join(tmp[1:], ""), nil
}

func hostedZoneComment(cluster awstpr.CustomObject) string {
	return fmt.Sprintf("Hosted zone for cluster %s", cluster.Spec.Cluster.Cluster.ID)
}
