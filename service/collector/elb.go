package collector

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
)

const (
	labelELB = "elb"
	// maxELBsInOneDescribeTagsBatch - https://docs.aws.amazon.com/elasticloadbalancing/2012-06-01/APIReference/API_DescribeTags.html
	maxELBsInOneDescribeTagsBatch = 20
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
		// AWS API has a limit for maximum number of LoadBalancerNames in
		// single Describe request so it must be done in batches of
		// maxELBsInOneBatch. In order to not spend so much time on this,
		// perform requests concurrently and synchronize them with errgroup.
		errGroup, _ := errgroup.WithContext(context.TODO())
		// Allocate list for ELB tag description results. Length is number of
		// AWS API batches + 1 for possible left over from uneven number.
		tagOutputs := make([]*elb.DescribeTagsOutput, len(loadBalancers)/maxELBsInOneDescribeTagsBatch+1)

		i := 0
		tagInput := &elb.DescribeTagsInput{}
		for _, lb := range loadBalancers {
			if len(tagInput.LoadBalancerNames) == maxELBsInOneDescribeTagsBatch {
				copyTagInput := tagInput
				errGroup.Go(func() error {
					o, err := awsClients.ELB.DescribeTags(copyTagInput)
					if err != nil {
						return microerror.Mask(err)
					}
					tagOutputs[i] = o
					return nil
				})

				tagInput = &elb.DescribeTagsInput{}
				i++
			}

			tagInput.LoadBalancerNames = append(tagInput.LoadBalancerNames, lb.LoadBalancerName)
		}

		// Last batch.
		if len(tagInput.LoadBalancerNames) > 0 {
			errGroup.Go(func() error {
				o, err := awsClients.ELB.DescribeTags(tagInput)
				if err != nil {
					return microerror.Mask(err)
				}
				tagOutputs[i] = o
				return nil
			})
		}

		// Now wait for all requests to complete.
		err := errGroup.Wait()
		if err != nil {
			return microerror.Mask(err)
		}

		// Extract tags from responses and create further loadBalancer types
		// based on these.
		for _, o := range tagOutputs {
			if o == nil {
				continue
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
