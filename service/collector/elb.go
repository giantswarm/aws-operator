package collector

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
)

const (
	AccountLabel = "account"
)

var (
	elbsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "elb_instance_out_of_service_count"),
		"Gauge about ELB instances being out of service.",
		[]string{
			AccountLabel,
			ClusterLabel,
			InstallationLabel,
			OrganizationLabel,
		},
		nil,
	)
)

func (c *Collector) collectAccountsELBs(ch chan<- prometheus.Metric, clients []aws.Clients) {
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go func(awsClients aws.Clients) {
			defer wg.Done()
			err := c.collectELBs(ch, awsClients)
			if err != nil {
				c.logger.Log("level", "error", "message", "failed collecting ELB metrics", "stack", fmt.Sprintf("%#v", err))
			}
		}(client)
	}

	wg.Wait()
}

func (c *Collector) collectELBs(ch chan<- prometheus.Metric, awsClients aws.Clients) error {
	c.logger.Log("level", "debug", "message", "collecting metrics for elbs")

	account, err := c.awsAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	var loadbalancers []*elb.LoadBalancerDescription
	{
		i := &elb.DescribeLoadBalancersInput{}
		o, err := awsClients.ELB.DescribeLoadBalancers(i)
		if err != nil {
			return microerror.Mask(err)
		}
		loadbalancers = o.LoadBalancerDescriptions
	}

	for _, l := range loadbalancers {
		var tags []*elb.Tag
		{
			i := &elb.DescribeTagsInput{
				LoadBalancerNames: []*string{
					l.LoadBalancerName,
				},
			}

			o, err := awsClients.ELB.DescribeTags(i)
			if err != nil {
				return microerror.Mask(err)
			}
			for _, d := range o.TagDescriptions {
				tags = append(tags, d.Tags...)
			}
		}

		var cluster string
		var installation string
		var organization string
		for _, t := range tags {
			if *t.Key == ClusterTag {
				cluster = *t.Value
			}
			if *t.Key == InstallationTag {
				installation = *t.Value
			}
			if *t.Key == OrganizationTag {
				organization = *t.Value
			}
		}

		if installation != c.installationName {
			continue
		}

		var count float64
		{
			i := &elb.DescribeInstanceHealthInput{
				Instances:        l.Instances,
				LoadBalancerName: l.LoadBalancerName,
			}

			o, err := awsClients.ELB.DescribeInstanceHealth(i)
			if err != nil {
				return microerror.Mask(err)
			}
			for _, s := range o.InstanceStates {
				if *s.State == "OutOfService" {
					count++
				}
			}
		}

		ch <- prometheus.MustNewConstMetric(
			elbsDesc,
			prometheus.GaugeValue,
			count,
			account,
			cluster,
			installation,
			organization,
		)
	}

	c.logger.Log("level", "debug", "message", "collected metrics for elbs")

	return nil
}
