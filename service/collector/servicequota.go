package collector

import (
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	labelServiceQuota = "service_quota"
)

const (
	subsystemServiceQuota = "servicequota"
)

var (
	serviceQuotaDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemServiceQuota, "info"),
		"Service Quota information.",
		[]string{
			labelAccountID,
			labelServiceQuota,
		},
		nil,
	)
	NATQuotaCode = "L-FE5A380F"
)

type ServiceQuotaConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type ServiceQuota struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

func NewServiceQuota(config ServiceQuotaConfig) (*ServiceQuota, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	v := &ServiceQuota{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return v, nil
}

func (v *ServiceQuota) Collect(ch chan<- prometheus.Metric) error {
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

func (v *ServiceQuota) Describe(ch chan<- *prometheus.Desc) error {
	ch <- serviceQuotaDesc
	return nil
}

func (v *ServiceQuota) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	var NATQuotaValue float64
	id := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   &NATQuotaCode,
		ServiceCode: &VPCServiceCode,
	}
	od, err := awsClients.ServiceQuotas.GetAWSDefaultServiceQuota(id)
	if err != nil {
		return microerror.Mask(err)
	}
	NATQuotaValue = *od.Quota.Value

	accountID, err := v.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	il := &servicequotas.ListServiceQuotasInput{
		ServiceCode: &VPCServiceCode,
	}
	ol, err := awsClients.ServiceQuotas.ListServiceQuotas(il)
	if err != nil {
		return microerror.Mask(err)
	}
	for _, sq := range ol.Quotas {
		if *sq.QuotaCode == NATQuotaCode {
			NATQuotaValue = *sq.Value
		}
	}

	ch <- prometheus.MustNewConstMetric(
		serviceQuotaDesc,
		prometheus.GaugeValue,
		NATQuotaValue,
		accountID,
		*od.Quota.QuotaName,
	)

	return nil
}
