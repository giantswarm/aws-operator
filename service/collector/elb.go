package collector

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cache"
)

const (
	// __ELBCache__ is used as temporal cache key to save ELB response.
	prefixELBcacheKey = "__ELBCache__"
	labelELB          = "elb"
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
	cache  *elbCache
	helper *helper
	logger micrologger.Logger

	installationName string
}

type elbCache struct {
	cache *cache.StringCache
}

type elbInfoResponse struct {
	Elbs []elbInfo
}

type elbInfo struct {
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
		cache:  newELBCache(time.Minute * 5),
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return e, nil
}

func newELBCache(expiration time.Duration) *elbCache {
	cache := &elbCache{
		cache: cache.NewStringCache(expiration),
	}

	return cache
}

func (n *elbCache) Get(key string) (*elbInfoResponse, error) {
	var c elbInfoResponse
	raw, exists := n.cache.Get(getELBCacheKey(key))
	if exists {
		err := json.Unmarshal(raw, &c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &c, nil
}

func (n *elbCache) Set(key string, content elbInfoResponse) error {
	contentSerialized, err := json.Marshal(content)
	if err != nil {
		return microerror.Mask(err)
	}

	n.cache.Set(getELBCacheKey(key), contentSerialized)

	return nil
}

func getELBCacheKey(key string) string {
	return prefixELBcacheKey + key
}

func (e *ELB) Collect(ch chan<- prometheus.Metric) error {
	reconciledClusters, err := e.helper.ListReconciledClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	awsClientsList, err := e.helper.GetAWSClients(reconciledClusters)
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

	var elbInfo *elbInfoResponse
	// Check if response is cached
	elbInfo, err = e.cache.Get(account)
	if err != nil {
		return microerror.Mask(err)
	}

	//Cache empty, getting from API
	if elbInfo == nil || elbInfo.Elbs == nil {
		elbInfo, err = getElbInfoFromAPI(account, e.installationName, awsClients)
		if err != nil {
			return microerror.Mask(err)
		}

		err = e.cache.Set(account, *elbInfo)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if elbInfo != nil {
		for _, lb := range elbInfo.Elbs {
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

// getElbInfoFromAPI collects ELB Info from AWS API
func getElbInfoFromAPI(account string, installation string, awsClients clientaws.Clients) (*elbInfoResponse, error) {
	var res elbInfoResponse

	var loadBalancerNames []*string
	{
		i := &elb.DescribeLoadBalancersInput{}
		o, err := awsClients.ELB.DescribeLoadBalancers(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		for _, d := range o.LoadBalancerDescriptions {
			loadBalancerNames = append(loadBalancerNames, d.LoadBalancerName)
		}

		if len(loadBalancerNames) == 0 {
			// E.g. during cluster creation there are no load balancers present
			// yet so further AWS API calls would fail on validation. No
			// metrics to emit either so we can short circuit here.
			return nil, nil
		}
	}

	var lbs []elbInfo
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
			return nil, microerror.Mask(err)
		}

		// Extract tags from responses and create further loadBalancer types
		// based on these.
		for _, o := range tagOutputs {
			for _, d := range o.TagDescriptions {
				lb := elbInfo{
					Name: *d.LoadBalancerName,
					Tags: make(map[string]string),
				}

				for _, t := range d.Tags {
					lb.Tags[*t.Key] = *t.Value
				}

				if lb.Tags[key.TagInstallation] != installation {
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
				return nil, microerror.Mask(err)
			}

			for _, s := range o.InstanceStates {
				if *s.State == stateOutOfService {
					lb.InstancesOutOfService++
				}
			}
		}
	}
	res.Elbs = lbs

	return &res, nil
}
