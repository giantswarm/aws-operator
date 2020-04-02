package collector

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	var loadBalancerNames []*string
	{
		i := &elb.DescribeLoadBalancersInput{}
		o, err := awsClients.ELB.DescribeLoadBalancers(i)
		if err != nil {
			return microerror.Mask(err)
		}
		for _, d := range o.LoadBalancerDescriptions {
			loadBalancerNames = append(loadBalancerNames, d.LoadBalancerName)
		}

		if len(loadBalancerNames) == 0 {
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
		// Slice for ELB tag description results.
		var tagOutputs []*elb.DescribeTagsOutput

		lbNames := loadBalancerNames
		mutex := &sync.Mutex{}
		for len(lbNames) > 0 {
			batchSize := maxELBsInOneDescribeTagsBatch
			if len(lbNames) < batchSize {
				batchSize = len(lbNames)
			}

			tagInput := &elb.DescribeTagsInput{
				LoadBalancerNames: lbNames[0:batchSize],
			}
			lbNames = lbNames[batchSize:]

			errGroup.Go(func() error {
				o, err := awsClients.ELB.DescribeTags(tagInput)
				if err != nil {
					return microerror.Mask(err)
				}

				mutex.Lock()
				tagOutputs = append(tagOutputs, o)
				mutex.Unlock()

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
			for _, d := range o.TagDescriptions {
				lb := loadBalancer{
					Name: *d.LoadBalancerName,
					Tags: make(map[string]string),
				}

				for _, t := range d.Tags {
					lb.Tags[*t.Key] = *t.Value
				}

				// Do not publish metrics for this cluster if it's version does not
				// match pkg/project/project.go version.
				ok, err := e.helper.IsClusterReconciledByThisVersion(lb.Tags[tagCluster])
				if err != nil {
					return microerror.Mask(err)
				}
				if !ok {
					continue
				}

				if lb.Tags[key.TagInstallation] != e.installationName {
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
				lb.Tags[key.TagInstallation],
				lb.Tags[tagOrganization],
			)
		}
	}

	return nil
}
