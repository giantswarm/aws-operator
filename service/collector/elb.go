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
	labelELB = "elb"
)

const (
	subsystemELB = "elb"
)

const (
	stateOutOfService = "OutOfService"
)

var (
	elbsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemELB, "instance_out_of_service_count"),
		"Gauge about ELB instances being out of service.",
		[]string{
			labelELB,
			labelAccount,
			labelCluster,
			labelInstallation,
			labelOrganization,
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

type loadBalancer struct {
	InstancesOutOfService float64
	Name                  string
	Tags                  map[string]string
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

	var loadBalancers []*elb.LoadBalancerDescription
	{
		i := &elb.DescribeLoadBalancersInput{}
		o, err := awsClients.ELB.DescribeLoadBalancers(i)
		if err != nil {
			return microerror.Mask(err)
		}
		loadBalancers = o.LoadBalancerDescriptions

		if len(loadBalancers) == 0 {
			// E.g. during cluster creation there are no load balancers present
			// yet so further AWS API calls would fail on validation. No
			// metrics to emit either so we can short circuit here.
			return nil
		}
	}

	var lbs []loadBalancer
	{
		i := &elb.DescribeTagsInput{}
		for _, l := range loadBalancers {
			i.LoadBalancerNames = append(i.LoadBalancerNames, l.LoadBalancerName)
		}

		o, err := awsClients.ELB.DescribeTags(i)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, d := range o.TagDescriptions {
			lb := loadBalancer{
				Name: *d.LoadBalancerName,
				Tags: make(map[string]string),
			}

			for _, t := range d.Tags {
				lb.Tags[*t.Key] = *t.Value
			}

			if lb.Tags[tagInstallation] != e.installationName {
				continue
			}

			lbs = append(lbs, lb)
		}
	}

	{
		// AWS API doesn't provide a method to describe instance health for all
		// specified ELBs so it must be done with N API calls.
		for _, lb := range lbs {
			i := &elb.DescribeInstanceHealthInput{
				LoadBalancerName: &lb.Name,
			}

			o, err := awsClients.ELB.DescribeInstanceHealth(i)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, s := range o.InstanceStates {
				if *s.State == stateOutOfService {
					lb.InstancesOutOfService++
				}
			}
		}
	}

	{
		for _, lb := range lbs {
			ch <- prometheus.MustNewConstMetric(
				elbsDesc,
				prometheus.GaugeValue,
				lb.InstancesOutOfService,
				lb.Name,
				account,
				lb.Tags[tagCluster],
				lb.Tags[tagInstallation],
				lb.Tags[tagOrganization],
			)
		}
	}

	return nil
}
