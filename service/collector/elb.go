package collector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	AccountLabel = "account"
	ELBLabel     = "elb"
)

const (
	StateOutOfService = "OutOfService"
)

var (
	elbsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "elb_instance_out_of_service_count"),
		"Gauge about ELB instances being out of service.",
		[]string{
			ELBLabel,
			AccountLabel,
			ClusterLabel,
			InstallationLabel,
			OrganizationLabel,
		},
		nil,
	)
)

type ELBConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type ELBCollector struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

func NewELB(config ELBConfig) (*ELBCollector, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	c := &ELBCollector{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return c, nil
}

func (c *ELBCollector) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := c.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("\n")
	fmt.Printf("len(awsClientsList): %#v\n", len(awsClientsList))
	fmt.Printf("\n")

	fmt.Printf("\n")
	fmt.Printf("start collect\n")
	fmt.Printf("\n")

	var g errgroup.Group

	for _, awsClients := range awsClientsList {
		fmt.Printf("1\n")
		g.Go(func() error {
			fmt.Printf("2\n")
			err := c.collectForAccount(ch, awsClients)
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

	fmt.Printf("\n")
	fmt.Printf("end collect\n")
	fmt.Printf("\n")

	return nil
}

func (c *ELBCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- elbsDesc
	return nil
}

func (c *ELBCollector) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	account, err := c.helper.AWSAccountID(awsClients)
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

	fmt.Printf("\n")
	fmt.Printf("start loop\n")
	fmt.Printf("\n")

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
				if *s.State == StateOutOfService {
					count++
				}
			}
		}

		fmt.Printf("\n")
		fmt.Printf("\n")
		fmt.Printf("\n")
		fmt.Printf("  count: %#v\n", count)
		fmt.Printf("  *l.LoadBalancerName: %#v\n", *l.LoadBalancerName)
		fmt.Printf("  account: %#v\n", account)
		fmt.Printf("  cluster: %#v\n", cluster)
		fmt.Printf("  installation: %#v\n", installation)
		fmt.Printf("  organization: %#v\n", organization)
		fmt.Printf("\n")
		fmt.Printf("\n")
		fmt.Printf("\n")

		ch <- prometheus.MustNewConstMetric(
			elbsDesc,
			prometheus.GaugeValue,
			count,
			*l.LoadBalancerName,
			account,
			cluster,
			installation,
			organization,
		)
	}

	fmt.Printf("\n")
	fmt.Printf("end loop\n")
	fmt.Printf("\n")

	return nil
}
