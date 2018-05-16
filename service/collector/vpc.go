package collector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ClusterTag      = "giantswarm.io/cluster"
	InstallationTag = "giantswarm.io/installation"
	NameTag         = "Name"
	OrganizationTag = "giantswarm.io/organization"
	StackNameTag    = "aws:cloudformation:stack-name"

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
	vpcs *prometheus.Desc = prometheus.NewDesc(
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

	f := func(filters []*ec2.Filter) {
		i := &ec2.DescribeVpcsInput{
			Filters: filters,
		}
		o, err := c.awsClients.EC2.DescribeVpcs(i)
		if err != nil {
			c.logger.Log("level", "error", "message", "could not list vpcs", "stack", fmt.Sprintf("%#v", err))
		}

		for _, vpc := range o.Vpcs {
			cluster := ""
			installation := ""
			name := ""
			organization := ""
			stackName := ""

			fmt.Printf("\n")
			fmt.Printf("\n")
			fmt.Printf("\n")
			for _, filter := range filters {
				fmt.Printf("%#v\n", filter.String())
				fmt.Printf("%#v\n", *filter.Name)
				fmt.Printf("%#v\n", *filter.Name)
				for _, v := range filter.Values {
					fmt.Printf("%#v\n", *v)
				}
			}
			fmt.Printf("c.installationName: %#v\n", c.installationName)
			fmt.Printf("vpc.ID: %#v\n", *vpc.VpcId)
			for _, tag := range vpc.Tags {
				fmt.Printf("tag: %#v\n", tag)
				fmt.Printf("tag.key: %#v\n", *tag.Key)
				fmt.Printf("tag.val: %#v\n", *tag.Value)
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
				vpcs,
				prometheus.GaugeValue,
				gaugeValue,
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
	}

	{
		filters := []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyInstallation)),
				Values: []*string{
					aws.String(c.installationName),
				},
			},
		}
		f(filters)
	}

	// TODO this is the deprecated tag we are only still using for old
	// clusters. This filter condition should be removed at some point.
	{
		filters := []*ec2.Filter{
			{
				Name: aws.String("tag:Installation"),
				Values: []*string{
					aws.String(c.installationName),
				},
			},
		}
		f(filters)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for vpcs")
}
