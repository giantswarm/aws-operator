package collector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Namespace = "aws_operator"

	ClusterTag      = "giantswarm.io/cluster"
	InstallationTag = "giantswarm.io/installation"
	NameTag         = "Name"
	OrganizationTag = "giantswarm.io/organization"
	StackNameTag    = "aws:cloudformation:stack-name"

	GaugeValue float64 = 1

	CidrLabel         = "cidr"
	ClusterLabel      = "cluster_id"
	IDLabel           = "id"
	InstallationLabel = "installation"
	NameLabel         = "name"
	OrganizationLabel = "organization"
	StackNameLabel    = "stack_name"
	StateLabel        = "state"
)

var (
	vpcsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "vpc_info"),
		"VPC information.",
		[]string{
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

func (c *Collector) collectVPCs(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics for vpcs")

	i := &ec2.DescribeVpcsInput{}
	o, err := c.awsClients.EC2.DescribeVpcs(i)
	if err != nil {
		c.logger.Log("level", "error", "message", "could not list vpcs", "stack", fmt.Sprintf("%#v", err))
	}

	for _, vpc := range o.Vpcs {
		installationName := installationFromTags(vpc.Tags)

		c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' belongs to installation '%s'", *vpc.VpcId, installationName))
		if installationName != c.installationName {
			c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' is being skipped for metrics collection", *vpc.VpcId))
			continue
		}
		c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' is being used for metrics collection", *vpc.VpcId))

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

		ch <- prometheus.MustNewConstMetric(
			vpcsDesc,
			prometheus.GaugeValue,
			GaugeValue,
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

func installationFromTags(tags []*ec2.Tag) string {
	for _, t := range tags {
		if *t.Key == InstallationTag {
			return *t.Value
		}

		// TODO the old hard coded tag "Installation" should be removed at some
		// point. Then we can get rid of this extra check.
		if *t.Key == "Installation" {
			return *t.Value
		}
	}

	return ""
}
