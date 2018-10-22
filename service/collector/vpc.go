package collector

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/aws-operator/client/aws"
)

const (
	NameTag      = "Name"
	StackNameTag = "aws:cloudformation:stack-name"

	AccountIdLabel = "account_id"
	CidrLabel      = "cidr"
	IDLabel        = "id"
	NameLabel      = "name"
	StackNameLabel = "stack_name"
	StateLabel     = "state"
)

var (
	vpcsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "vpc_info"),
		"VPC information.",
		[]string{
			AccountIdLabel,
			CidrLabel,
			ClusterLabel,
			IDLabel,
			InstallationLabel,
			NameLabel,
			OrganizationLabel,
			StackNameLabel,
			StateLabel,
		},
		nil,
	)
)

func (c *Collector) collectAccountsVPCs(ch chan<- prometheus.Metric, clients []aws.Clients) {
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go func(awsClients aws.Clients) {
			defer wg.Done()
			c.collectVPCs(ch, awsClients)
		}(client)
	}

	wg.Wait()
}

func (c *Collector) collectVPCs(ch chan<- prometheus.Metric, awsClients aws.Clients) {
	c.logger.Log("level", "debug", "message", "collecting metrics for vpcs")

	i := &ec2.DescribeVpcsInput{}
	o, err := awsClients.EC2.DescribeVpcs(i)
	if err != nil {
		c.logger.Log("level", "error", "message", "could not list vpcs", "stack", fmt.Sprintf("%#v", err))
	}

	accountID, err := c.awsAccountID(awsClients)
	if err != nil {
		c.logger.Log("level", "error", "message", "could not get aws account id", "stack", fmt.Sprintf("%#v", err))
	}

	for _, vpc := range o.Vpcs {
		cluster := ""
		installation := ""
		name := ""
		organization := ""
		stackName := ""

		for _, tag := range vpc.Tags {
			if *tag.Key == ClusterTag {
				cluster = *tag.Value
			}
			if *tag.Key == InstallationTag {
				installation = *tag.Value
			}
			if *tag.Key == NameTag {
				name = *tag.Value
			}
			if *tag.Key == OrganizationTag {
				organization = *tag.Value
			}
			if *tag.Key == StackNameTag {
				stackName = *tag.Value
			}
		}

		c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' belongs to installation '%s'", *vpc.VpcId, installation))
		if installation != c.installationName {
			c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' is being skipped for metrics collection", *vpc.VpcId))
			continue
		}
		c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' is being used for metrics collection", *vpc.VpcId))

		ch <- prometheus.MustNewConstMetric(
			vpcsDesc,
			prometheus.GaugeValue,
			GaugeValue,
			accountID,
			*vpc.CidrBlock,
			cluster,
			*vpc.VpcId,
			installation,
			name,
			organization,
			stackName,
			*vpc.State,
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for vpcs")
}
