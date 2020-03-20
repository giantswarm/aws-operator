package collector

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cache"
)

const (
	//awsNATlocker is used as temporal lock to decide if AWS NAT API request should be done or not
	awsNATlocker = "__awsNATlocker__"
	labelVPC     = "vpc"
	labelAZ      = "availability_zone"
)

const (
	subsystemNAT = "nat"
)

var (
	natDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemNAT, "info"),
		"NAT limits information.",
		[]string{
			labelAccountID,
			labelVPC,
			labelAZ,
		},
		nil,
	)
)

type NATConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type NAT struct {
	awsAPIcache *cache.Float64Cache
	helper      *helper
	logger      micrologger.Logger

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
		//AWS operator creates at this moment one NAT for each private subnet (node pool).
		//As clusters are not created or changed so often, and the process can take around 20 minutes,
		//30 minutes for the cache expiration is a coherent value.
		awsAPIcache: cache.NewFloat64Cache(time.Minute * 30),
		helper:      config.Helper,
		logger:      config.Logger,

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
	if _, ok := v.awsAPIcache.Get(awsNATlocker); ok {
		return nil
	}
	v.awsAPIcache.Set(awsNATlocker, 1)

	accountID, err := v.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	iv := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(key.TagOrganization),
				Values: []*string{
					aws.String("giantswarm"),
				},
			},
		},
	}
	rv, err := awsClients.EC2.DescribeVpcs(iv)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, vpc := range rv.Vpcs {
		in := &ec2.DescribeNatGatewaysInput{
			Filter: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						vpc.VpcId,
					},
				},
			},
		}

		var azs map[string]float64
		azs = make(map[string]float64)
		rn, err := awsClients.EC2.DescribeNatGateways(in)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, nat := range rn.NatGateways {
			is := &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{
					nat.SubnetId,
				},
			}
			rs, err := awsClients.EC2.DescribeSubnets(is)
			if err != nil {
				return microerror.Mask(err)
			}
			for _, sub := range rs.Subnets {
				zoneID := *sub.AvailabilityZoneId
				if _, ok := azs[zoneID]; ok {
					azs[zoneID] = azs[zoneID] + 1
				} else {
					azs[zoneID] = 1
				}
			}
		}

		for azName, azValue := range azs {
			ch <- prometheus.MustNewConstMetric(
				natDesc,
				prometheus.GaugeValue,
				azValue,
				accountID,
				*vpc.VpcId,
				azName,
			)
		}

	}

	return nil
}
