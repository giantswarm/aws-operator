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
		c.logger.Log("level", "debug", "message", fmt.Sprintf("checking if VPC '%s' belongs to current installation", *vpc.VpcId))
		if !containsInstallationTag(vpc.Tags, c.installationName) {
			c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' does not belong to current installation", *vpc.VpcId))
			continue
		}
		c.logger.Log("level", "debug", "message", fmt.Sprintf("VPC '%s' belongs to current installation", *vpc.VpcId))

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

		fmt.Printf("emitting metric for %#v\n", vpc)

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

func containsInstallationTag(tags []*ec2.Tag, n string) bool {
	for _, t := range tags {
		if *t.Key != InstallationTag {
			continue
		}
		// TODO this is the old tag which should be removed at some point.
		if *t.Key != "Installation" {
			continue
		}
		if *t.Value == n {
			return true
		}
	}

	return false
}
