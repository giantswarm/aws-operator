package create

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
)

type recordSetInput struct {
	Cluster      awstpr.CustomObject
	Client       *route53.Route53
	Resource     resources.DNSNamedResource
	Value        string
	Domain       string
	HostedZoneID string
	Type         string
}

type hostedZoneInput struct {
	Cluster   awstpr.CustomObject
	Domain    string
	Private   bool
	Client    *route53.Route53
	VPCID     string
	VPCRegion string
}

func (s *Service) deleteRecordSet(input recordSetInput) error {
	rs := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Value:        input.Value,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
		Type:         input.Type,
	}

	if err := rs.Delete(); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (s *Service) createRecordSet(input recordSetInput) error {
	// Create DNS records for LB.
	apiRecordSet := &awsresources.RecordSet{
		Client:       input.Client,
		Resource:     input.Resource,
		Value:        input.Value,
		Domain:       input.Domain,
		HostedZoneID: input.HostedZoneID,
		Type:         input.Type,
	}

	if err := apiRecordSet.CreateOrFail(); err != nil {
		return microerror.Maskf(err, "error registering DNS record '%s'", apiRecordSet.Domain)
	}

	s.logger.Log("debug", fmt.Sprintf("created or reused DNS record '%s'", apiRecordSet.Domain))

	return nil
}

func (s *Service) createHostedZone(input hostedZoneInput) (*awsresources.HostedZone, error) {
	hzName, err := hostedZoneName(input.Domain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	hz := &awsresources.HostedZone{
		Name:      hzName,
		Comment:   hostedZoneComment(input.Cluster),
		Private:   input.Private,
		Client:    input.Client,
		VPCID:     input.VPCID,
		VPCRegion: input.VPCRegion,
	}

	hzCreated, err := hz.CreateIfNotExists()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if hzCreated {
		s.logger.Log("debug", fmt.Sprintf("created hosted zone '%s'", hz.Name))
	} else {
		s.logger.Log("debug", fmt.Sprintf("hosted zone '%s' already exists, reusing", hz.Name))
	}

	return hz, nil
}

func (s *Service) deleteHostedZone(input hostedZoneInput) error {
	hzName, err := hostedZoneName(input.Domain)
	if err != nil {
		return microerror.Mask(err)
	}

	hz := &awsresources.HostedZone{
		Name: hzName,
	}

	err = hz.Delete()
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func hostedZoneComment(cluster awstpr.CustomObject) string {
	return fmt.Sprintf("Hosted zone for cluster %s", cluster.Spec.Cluster.Cluster.ID)
}

// hostedZoneName removes the first 2 subdomains from the domain
// e.g.  apiserver.foobar.aws.giantswarm.io -> aws.giantswarm.io
func hostedZoneName(domain string) (string, error) {
	tmp := strings.SplitN(domain, ".", 3)

	if len(tmp) != 3 {
		return "", microerror.Mask(malformedCloudConfigKeyError)
	}

	return strings.Join(tmp[2:], ""), nil
}
