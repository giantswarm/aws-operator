package collector

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/internal/cache"
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
	NATQuotaCode   = "L-FE5A380F"
	NATQuotaName   = "nat-gateway"
	VPCServiceCode = "vpc"
)

type ServiceQuotaConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type ServiceQuota struct {
	awsAPIcache *cache.Float64Cache
	helper      *helper
	logger      micrologger.Logger

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
		// Default quotas are changed by request to AWS support and they are
		// considered quite static information, then 12 hours for the cache
		// expiration is a reasonable value.
		awsAPIcache: cache.NewFloat64Cache(time.Minute * 720),
		helper:      config.Helper,
		logger:      config.Logger,

		installationName: config.InstallationName,
	}

	return v, nil
}

func (v *ServiceQuota) Collect(ch chan<- prometheus.Metric) error {
	reconciledClusters, err := v.helper.ListReconciledClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	awsClientsList, err := v.helper.GetAWSClients(context.Background(), reconciledClusters)
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
	accountID, err := v.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	// natQuotaValue reflects the value of number of NAT Gateways that can be
	// created by the operator in a specific VPC for each availability zone.
	var natQuotaValue float64
	if val, ok := v.awsAPIcache.Get(NATQuotaCode); ok {
		natQuotaValue = val
	} else {
		natQuotaValue, err := getDefaultVPCQuotaFor(NATQuotaCode, awsClients)
		if err != nil {
			return microerror.Mask(err)
		}
		v.awsAPIcache.Set(NATQuotaCode, natQuotaValue)
	}

	ch <- prometheus.MustNewConstMetric(
		serviceQuotaDesc,
		prometheus.GaugeValue,
		natQuotaValue,
		accountID,
		NATQuotaName,
	)

	return nil
}

func getDefaultVPCQuotaFor(quotaCode string, awsClients clientaws.Clients) (float64, error) {
	id := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   &NATQuotaCode,
		ServiceCode: &VPCServiceCode,
	}
	//Get the default NAT quota for the specific account
	od, err := awsClients.ServiceQuotas.GetAWSDefaultServiceQuota(id)
	if IsEndpointNotAvailable(err) {
		// Some regions do not support ServiceQuota API.
		return 0, nil
	} else if err != nil {
		return 0, microerror.Mask(err)
	}
	natQuotaValue := *od.Quota.Value

	il := &servicequotas.ListServiceQuotasInput{
		ServiceCode: &VPCServiceCode,
	}
	//Get the NAT quota in case it has been modified by AWS support request
	ol, err := awsClients.ServiceQuotas.ListServiceQuotas(il)
	if err != nil {
		return natQuotaValue, microerror.Mask(err)
	}
	for _, sq := range ol.Quotas {
		if *sq.QuotaCode == NATQuotaCode {
			natQuotaValue = *sq.Value
		}
	}

	return natQuotaValue, nil
}
