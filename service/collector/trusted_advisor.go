package collector

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/service/support"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/giantswarm/aws-operator/client/aws"
)

const (
	// serviceLimitCategory is the category returned by Trusted Advisor
	// for checks related to service limits and usage.
	serviceLimitCategory = "service_limits"

	// resourceMetadataLength is the length of resource metadata we expect.
	resourceMetadataLength = 6

	regionIndex    = 0
	serviceIndex   = 1
	limitNameIndex = 2
	limitIndex     = 3
	usageIndex     = 4

	accountIDLabel = "account_id"
	regionLabel    = "region"
	serviceLabel   = "service"
	limitNameLabel = "name"
)

var (
	serviceLimit *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "service_limit"),
		"Service limits as reported by Trusted Advisor.",
		[]string{
			accountIDLabel,
			regionLabel,
			serviceLabel,
			limitNameLabel,
		},
		nil,
	)
	serviceUsage *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "service_usage"),
		"Service usage as reported by Trusted Advisor.",
		[]string{
			accountIDLabel,
			regionLabel,
			serviceLabel,
			limitNameLabel,
		},
		nil,
	)

	trustedAdvisorSupport *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "trusted_advisor_supported"),
		"Gauge describing whether Trusted Advisor is available with your support plan.",
		nil, nil,
	)

	trustedAdvisorError = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_error_count",
		Help:      "Counter for the number of errors encountered calling Trusted Advisor.",
	})

	getChecksDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_get_checks_duration",
		Help:      "Histogram for the duration of Trusted Advisor get checks calls.",
	})
	getResourcesDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_get_resources_duration",
		Help:      "Histogram for the duration of Trusted Advisor get resource calls.",
	})
)

func init() {
	prometheus.MustRegister(trustedAdvisorError)

	prometheus.MustRegister(getChecksDuration)
	prometheus.MustRegister(getResourcesDuration)
}

func (c *Collector) collectAccountsTrustedAdvisorChecks(ch chan<- prometheus.Metric, clients []aws.Clients) {
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go func(awsClients aws.Clients) {
			defer wg.Done()
			c.collectTrustedAdvisorChecks(ch, awsClients)
		}(client)
	}

	wg.Wait()
}

func (c *Collector) collectTrustedAdvisorChecks(ch chan<- prometheus.Metric, awsClients aws.Clients) {
	c.logger.Log("level", "debug", "message", "collecting metrics for trusted advisor checks")

	checks, err := c.getTrustedAdvisorChecks(awsClients)
	if err != nil {
		if IsUnsupportedPlan(err) {
			c.logger.Log("level", "info", "message", "Trusted Advisor not available in support plan, cannot fetch service usage metrics")
			ch <- trustedAdvisorNotSupported()
			return
		}

		c.logger.Log("level", "error", "message", "could not get Trusted Advisor checks", "stack", fmt.Sprintf("%#v", err))
		trustedAdvisorError.Inc()
		return
	}
	ch <- trustedAdvisorSupported()

	accountID, err := c.awsAccountID(awsClients)
	if err != nil {
		c.logger.Log("level", "error", "message", "could not get aws account id", "stack", fmt.Sprintf("%#v", err))
	}

	var wg sync.WaitGroup

	for _, check := range checks {
		// Ignore any checks that don't relate to service limits.
		if *check.Category != serviceLimitCategory {
			continue
		}

		wg.Add(1)

		go func(id *string) {
			defer wg.Done()

			resources, err := c.getTrustedAdvisorResources(id, awsClients)
			if err != nil {
				c.logger.Log("level", "error", "message", "could not get Trusted Advisor resource", "stack", fmt.Sprintf("%#v", err), "id", *id)
				trustedAdvisorError.Inc()
				return
			}

			for _, resource := range resources {
				// One Trusted Advisor check returns the nil string for current usage. Skip it.
				if len(resource.Metadata) == 6 && resource.Metadata[4] == nil {
					continue
				}

				limit, usage, err := resourceToMetrics(resource, accountID)
				if err != nil {
					c.logger.Log("level", "error", "message", "could not convert Trusted Advisor resource into metrics", "stack", fmt.Sprintf("%#v", err), "id", *id)
					trustedAdvisorError.Inc()
					return
				}

				ch <- limit
				ch <- usage
			}
		}(check.Id)
	}

	wg.Wait()

	c.logger.Log("level", "debug", "message", "finished collecting metrics for trusted advisor checks")
}

// getTrustedAdvisorCheckDescriptions calls Trusted Advisor API to get all available checks.
func (c *Collector) getTrustedAdvisorChecks(awsClients aws.Clients) ([]*support.TrustedAdvisorCheckDescription, error) {
	timer := prometheus.NewTimer(getChecksDuration)

	englishLanguage := "en"
	describeChecksInput := &support.DescribeTrustedAdvisorChecksInput{
		Language: &englishLanguage,
	}
	describeChecksOutput, err := awsClients.Support.DescribeTrustedAdvisorChecks(describeChecksInput)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	timer.ObserveDuration()

	return describeChecksOutput.Checks, nil
}

// getTrustedAdvisorResources calls Trusted Advisor API to get flagged resources of the given check ID.
func (c *Collector) getTrustedAdvisorResources(id *string, awsClients aws.Clients) ([]*support.TrustedAdvisorResourceDetail, error) {
	timer := prometheus.NewTimer(getResourcesDuration)

	checkResultInput := &support.DescribeTrustedAdvisorCheckResultInput{
		CheckId: id,
	}
	checkResultOutput, err := awsClients.Support.DescribeTrustedAdvisorCheckResult(checkResultInput)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	timer.ObserveDuration()

	return checkResultOutput.Result.FlaggedResources, nil
}

func resourceToMetrics(resource *support.TrustedAdvisorResourceDetail, accountID string) (prometheus.Metric, prometheus.Metric, error) {
	if len(resource.Metadata) != resourceMetadataLength {
		return nil, nil, invalidResourceError
	}

	region := resource.Metadata[regionIndex]
	service := resource.Metadata[serviceIndex]
	limitName := resource.Metadata[limitNameIndex]

	limit := resource.Metadata[limitIndex]
	usage := resource.Metadata[usageIndex]

	if limit == nil {
		return nil, nil, nilLimitError
	}
	if usage == nil {
		return nil, nil, nilUsageError
	}

	limitInt, err := strconv.Atoi(*limit)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	usageInt, err := strconv.Atoi(*usage)
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	limitMetric := prometheus.MustNewConstMetric(
		serviceLimit, prometheus.GaugeValue, float64(limitInt), accountID, *region, *service, *limitName,
	)
	usageMetric := prometheus.MustNewConstMetric(
		serviceUsage, prometheus.GaugeValue, float64(usageInt), accountID, *region, *service, *limitName,
	)

	return limitMetric, usageMetric, nil
}

func trustedAdvisorSupported() prometheus.Metric {
	return prometheus.MustNewConstMetric(
		trustedAdvisorSupport, prometheus.GaugeValue, float64(1),
	)
}

func trustedAdvisorNotSupported() prometheus.Metric {
	return prometheus.MustNewConstMetric(
		trustedAdvisorSupport, prometheus.GaugeValue, float64(0),
	)
}
