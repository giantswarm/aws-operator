package collector

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/hprose/hprose-go"
	"github.com/prometheus/client_golang/prometheus"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cache"
)

const (
	// __NATCache__ is used as temporal cache key to save NAT response.
	prefixNATcacheKey = "__NATCache__"
	labelVPC          = "vpc"
	labelAZ           = "availability_zone"
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
	cache  *natCache
	helper *helper
	logger micrologger.Logger

	installationName string
}

type natCache struct {
	cache *cache.StringCache
}

type natResponse struct {
	vpcs map[string]natVPC
}

type natVPC struct {
	zones map[string]float64
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
		// AWS operator creates at this moment one NAT for each private subnet (node
		// pool). As clusters are not created nor changed so often, and the process
		// can take around 20 minutes, 30 minutes for the cache expiration is a
		// reasonable value.
		cache: newNATCache(time.Minute * 5),
		//cache: newNATCache(time.Minute * 30),
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return v, nil
}

func newNATCache(expiration time.Duration) *natCache {
	cache := &natCache{
		cache: cache.NewStringCache(expiration),
	}

	return cache
}

func (n *natCache) Get(accID string) (*natResponse, error) {
	var c *natResponse

	raw, ok := n.cache.Get(prefixNATcacheKey + accID)
	if ok {
		err := hprose.Unserialize(raw, c, true)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return c, nil
}

func (n *natCache) Set(accID string, content *natResponse) error {
	contentSerialized, err := hprose.Serialize(content, true)
	if err != nil {
		return microerror.Mask(err)
	}

	n.cache.Set(prefixNATcacheKey+accID, contentSerialized)

	return nil
}

func (v *NAT) Collect(ch chan<- prometheus.Metric) error {
	reconciledClusters, err := v.helper.ListReconciledClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	awsClientsList, err := v.helper.GetAWSClients(reconciledClusters)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, item := range awsClientsList {
		awsClients := item

		err := v.collectForAccount(ch, awsClients)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (v *NAT) Describe(ch chan<- *prometheus.Desc) error {
	ch <- natDesc
	return nil
}

func (v *NAT) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	accountID, err := v.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Println("Collecting nat info")
	var natInfo *natResponse

	natInfo, err = v.cache.Get(accountID)
	if err != nil {
		return microerror.Mask(err)
	}

	//Cache empty, getting from API
	if natInfo == nil {
		natInfo, err := getNatInfoFromAPI(accountID, awsClients)
		if err != nil {
			return microerror.Mask(err)
		}
		err = v.cache.Set(accountID, natInfo)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for vpcID, vpcInfo := range natInfo.vpcs {
		for azName, azValue := range vpcInfo.zones {
			ch <- prometheus.MustNewConstMetric(
				natDesc,
				prometheus.GaugeValue,
				azValue,
				accountID,
				vpcID,
				azName,
			)
		}
	}

	return nil
}

func getNatInfoFromAPI(accountID string, awsClients clientaws.Clients) (*natResponse, error) {
	var res natResponse
	fmt.Printf("nat cache empty, query AWS API for account %s\n", accountID)

	iv := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", key.TagOrganization)),
				Values: []*string{
					aws.String("giantswarm"),
				},
			},
		},
	}
	rv, err := awsClients.EC2.DescribeVpcs(iv)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, vpc := range rv.Vpcs {

		fmt.Printf("nat vpc %s \n", *vpc.VpcId)
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

		res.vpcs[*vpc.VpcId] = natVPC{
			zones: make(map[string]float64),
		}

		rn, err := awsClients.EC2.DescribeNatGateways(in)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		for _, nat := range rn.NatGateways {
			fmt.Printf("nat subnet %s \n", *nat.SubnetId)
			is := &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{
					nat.SubnetId,
				},
			}
			rs, err := awsClients.EC2.DescribeSubnets(is)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			for _, sub := range rs.Subnets {
				zoneID := *sub.AvailabilityZoneId
				if _, ok := res.vpcs[*vpc.VpcId].zones[zoneID]; ok {
					res.vpcs[*vpc.VpcId].zones[zoneID] = res.vpcs[*vpc.VpcId].zones[zoneID] + 1
				} else {
					res.vpcs[*vpc.VpcId].zones[zoneID] = 1
				}
			}
		}

	}

	return &res, nil
}
