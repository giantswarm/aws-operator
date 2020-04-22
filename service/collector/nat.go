package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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

type natInfoResponse struct {
	Vpcs map[string]vpcInfo
}

type vpcInfo struct {
	NatGatewaysByZone map[string]float64
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
		cache:  newNATCache(time.Minute * 30),
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

func (n *natCache) Get(accID string) (*natInfoResponse, error) {
	var c natInfoResponse
	raw, exists := n.cache.Get(prefixNATcacheKey + accID)
	if exists {
		err := json.Unmarshal(raw, &c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &c, nil
}

func (n *natCache) Set(accID string, content natInfoResponse) error {
	contentSerialized, err := json.Marshal(content)
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

	var natInfo *natInfoResponse
	// Check if response is cached
	natInfo, err = v.cache.Get(accountID)
	if err != nil {
		return microerror.Mask(err)
	}

	//Cache empty, getting from API
	if natInfo == nil || natInfo.Vpcs == nil {
		natInfo, err = getNatInfoFromAPI(accountID, awsClients)
		if err != nil {
			return microerror.Mask(err)
		}

		err = v.cache.Set(accountID, *natInfo)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if natInfo != nil {
		for vpcID, vpcInfo := range natInfo.Vpcs {
			for azName, azValue := range vpcInfo.NatGatewaysByZone {
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
	}

	return nil
}

// getNatInfoFromAPI collect from AWS API the number of NAT Gateways by Availability Zone for
// each VPC of the installation
func getNatInfoFromAPI(accountID string, awsClients clientaws.Clients) (*natInfoResponse, error) {
	var res natInfoResponse
	res.Vpcs = make(map[string]vpcInfo)

	// 1. Get all VPCs of the installation
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

	// 2. Get all NAT GWs for each VPC
	for _, vpc := range rv.Vpcs {
		res.Vpcs[*vpc.VpcId] = vpcInfo{
			NatGatewaysByZone: make(map[string]float64),
		}

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
		rn, err := awsClients.EC2.DescribeNatGateways(in)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// 3. Get all the subnets for each NAT GW
		for _, nat := range rn.NatGateways {
			is := &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{
					nat.SubnetId,
				},
			}
			rs, err := awsClients.EC2.DescribeSubnets(is)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			// 4. Store the number of GWs by Availability Zone
			for _, sub := range rs.Subnets {
				vpcID := *vpc.VpcId
				zoneID := *sub.AvailabilityZoneId
				if _, exists := res.Vpcs[vpcID].NatGatewaysByZone[zoneID]; exists {
					res.Vpcs[vpcID].NatGatewaysByZone[zoneID] = res.Vpcs[vpcID].NatGatewaysByZone[zoneID] + 1
				} else {
					res.Vpcs[vpcID].NatGatewaysByZone[zoneID] = 1
				}
			}
		}

	}

	return &res, nil
}
