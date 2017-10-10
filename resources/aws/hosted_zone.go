package aws

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/giantswarm/microerror"
)

type HostedZone struct {
	Name      string
	id        string
	Private   bool
	Comment   string
	Client    *route53.Route53
	VPCID     string
	VPCRegion string
}

func (hz *HostedZone) CreateOrFail() error {
	callerReference := time.Now().UTC().String()

	createInput := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(callerReference),
		Name:            aws.String(hz.Name),
		HostedZoneConfig: &route53.HostedZoneConfig{
			Comment:     aws.String(hz.Comment),
			PrivateZone: aws.Bool(hz.Private),
		},
		VPC: &route53.VPC{
			VPCId:     aws.String(hz.VPCID),
			VPCRegion: aws.String(hz.VPCRegion),
		},
	}

	resp, err := hz.Client.CreateHostedZone(createInput)

	if err != nil {
		return microerror.Mask(err)
	}

	hz.id = *resp.HostedZone.Id

	return nil
}

func (hz *HostedZone) CreateIfNotExists() (bool, error) {
	exists, err := hz.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := hz.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (hz HostedZone) Delete() error {
	hostedZone, err := hz.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	if _, err := hz.Client.DeleteHostedZone(&route53.DeleteHostedZoneInput{
		Id: aws.String(*hostedZone.Id),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil

}

// NewHostedZoneFromExisting initializes a Hosted Zone, setting some fields it has retrieved from an existing HZ
// It's used when deleting a RecordSet. It does not create a new HZ on AWS.
func NewHostedZoneFromExisting(name string, client *route53.Route53) (*HostedZone, error) {
	hz := HostedZone{
		Name:   name,
		Client: client,
	}

	existingHz, err := hz.findExisting()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	hz.id = *existingHz.Id

	return &hz, nil
}

func (hz HostedZone) GetID() string {
	return hz.id
}

func (hz *HostedZone) findExisting() (*route53.HostedZone, error) {
	resp, err := hz.Client.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(hz.Name),
		MaxItems: aws.String("1"),
	})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	hostedZones := resp.HostedZones

	if len(hostedZones) == 0 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, HostedZoneType, hz.Name)
	}

	// this AWS endpoint returns all hosted zones, even ones that don't match the query
	// if there was a HZ that matched the DNSName, it will be the first one returned
	// so we need to match the first result by name
	hostedZone := hostedZones[0]

	// AWS returns the proper DNS name, i.e. with a trailing dot
	if strings.TrimRight(*hostedZone.Name, ".") != hz.Name {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, HostedZoneType, hz.Name)
	}

	return hostedZone, nil
}

func (hz *HostedZone) checkIfExists() (bool, error) {
	existingHz, err := hz.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	hz.id = *existingHz.Id

	return true, nil
}
