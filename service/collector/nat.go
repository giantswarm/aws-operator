package collector

import (
	//"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"

	"github.com/aws/aws-sdk-go/service/servicequotas"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
)

const (
	labelVPC = "vpc"
)

const (
	subsystemNAT = "nat"
)

var (
	natDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemNAT, "info"),
		"VPC information.",
		[]string{
			labelAccountID,
			labelCluster,
			labelInstallation,
			labelOrganization,
			labelVPC,
		},
		nil,
	)
	NATQuotaCode   string = "L-FE5A380F"
	VPCServiceCode string = "vpc"
)

type NATConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type NAT struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

func NewNAT(config NATConfig) (*NAT, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	v := &NAT{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return v, nil
}

func (v *NAT) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := v.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := v.collectForAccount(ch, awsClients)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (v *NAT) Describe(ch chan<- *prometheus.Desc) error {
	ch <- natDesc
	return nil
}

func (v *NAT) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	id := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   &NATQuotaCode,
		ServiceCode: &VPCServiceCode,
	}
	od, err := awsClients.ServiceQuotas.GetAWSDefaultServiceQuota(id)
	if err != nil {
		return microerror.Mask(err)
	}

	// accountID, err := v.helper.AWSAccountID(awsClients)
	// if err != nil {
	// 	return microerror.Mask(err)
	// }

	fmt.Printf("Service Quota NAT GW: %f", *od.Quota.Value)

	il := &servicequotas.ListServiceQuotasInput{
		ServiceCode: &VPCServiceCode,
	}
	ol, err := awsClients.ServiceQuotas.ListServiceQuotas(il)
	if err != nil {
		return microerror.Mask(err)
	}
	for _, sq := range ol.Quotas {
		fmt.Printf("Service quotas define %s: %f", *sq.QuotaName, *sq.Value)
	}

	return nil
}
