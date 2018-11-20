package collector

import (
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
)

const (
	ELBLabel = "elb"
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

type ELB struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

func NewELB(config ELBConfig) (*ELB, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	e := &ELB{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return e, nil
}

func (e *ELB) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := e.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := e.collectForAccount(ch, awsClients)
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

func (e *ELB) Describe(ch chan<- *prometheus.Desc) error {
	ch <- elbsDesc
	return nil
}

func (e *ELB) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	account, err := e.helper.AWSAccountID(awsClients)
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

		if installation != e.installationName {
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

	return nil
}
