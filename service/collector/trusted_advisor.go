package collector

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/support"
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// serviceLimitCategory is the category returned by Trusted Advisor
	// for checks related to service limits and usage.
	serviceLimitCategory = "service_limits"

	// trustedAdvisorUnsupportedErrorMessage is the error message returned
	// if Trusted Advisor is not supported (support plan is not Business or Enterprise).
	trustedAdvisorUnsupportedErrorMessage = "AWS Premium Support Subscription is required to use this service."

	regionIndex    = 0
	serviceIndex   = 1
	limitNameIndex = 2
	limitIndex     = 3
	usageIndex     = 4

	regionLabel    = "region"
	serviceLabel   = "service"
	limitNameLabel = "name"
)

var (
	serviceLimit *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "service_limit"),
		"Service limits as reported by Trusted Advisor.",
		[]string{
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

	getChecksError = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_get_checks_error_count",
		Help:      "Counter for the number of errors encountered getting Trusted Advisor checks.",
	})
	getResourcesError = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_get_resources_error_count",
		Help:      "Counter for the number of errors encountered getting Trusted Advisor resources.",
	})
	convertResourcesError = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "trusted_advisor_convert_resources_error_count",
		Help:      "Counter for the number of errors encountered converting Trusted Advisor resources.",
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
	prometheus.MustRegister(getChecksError)
	prometheus.MustRegister(getResourcesError)
	prometheus.MustRegister(convertResourcesError)

	prometheus.MustRegister(getChecksDuration)
	prometheus.MustRegister(getResourcesDuration)
}

func (c *Collector) collectTrustedAdvisorChecks(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics for trusted advisor checks")

	checks, err := c.getTrustedAdvisorChecks()
	if err != nil {
		if strings.Contains(err.Error(), trustedAdvisorUnsupportedErrorMessage) {
			c.logger.Log("level", "info", "message", "Trusted Advisor not available in support plan, cannot fetch service usage metrics")
			ch <- trustedAdvisorNotSupported()
			return
		}

		c.logger.Log("level", "error", "message", "could not get Trusted Advisor checks", "stack", fmt.Sprintf("%#v", err))
		getChecksError.Inc()
		return
	}
	ch <- trustedAdvisorSupported()

	var wg sync.WaitGroup

	for _, check := range checks {
		// Ignore any checks that don't relate to service limits.
		if *check.Category != serviceLimitCategory {
			continue
		}

		wg.Add(1)

		go func(id *string) {
			defer wg.Done()

			resources, err := c.getTrustedAdvisorResources(id)
			if err != nil {
				c.logger.Log("level", "error", "message", "could not get Trusted Advisor resource", "stack", fmt.Sprintf("%#v", err), "id", *id)
				getResourcesError.Inc()
				return
			}

			for _, resource := range resources {
				// One Trusted Advisor check returns the nil string for current usage. Skip it.
				if len(resource.Metadata) == 6 && resource.Metadata[4] == nil {
					continue
				}

				limit, usage, err := resourceToMetrics(resource)
				if err != nil {
					c.logger.Log("level", "error", "message", "could not convert Trusted Advisor resource into metrics", "stack", fmt.Sprintf("%#v", err), "id", *id)
					convertResourcesError.Inc()
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
func (c *Collector) getTrustedAdvisorChecks() ([]*support.TrustedAdvisorCheckDescription, error) {
	timer := prometheus.NewTimer(getChecksDuration)

	englishLanguage := "en"
	describeChecksInput := &support.DescribeTrustedAdvisorChecksInput{
		Language: &englishLanguage,
	}
	describeChecksOutput, err := c.awsClients.Support.DescribeTrustedAdvisorChecks(describeChecksInput)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	timer.ObserveDuration()

	return describeChecksOutput.Checks, nil
}

// getTrustedAdvisorResources calls Trusted Advisor API to get flagged resources of the given check ID.
func (c *Collector) getTrustedAdvisorResources(id *string) ([]*support.TrustedAdvisorResourceDetail, error) {
	timer := prometheus.NewTimer(getResourcesDuration)

	checkResultInput := &support.DescribeTrustedAdvisorCheckResultInput{
		CheckId: id,
	}
	checkResultOutput, err := c.awsClients.Support.DescribeTrustedAdvisorCheckResult(checkResultInput)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	timer.ObserveDuration()

	return checkResultOutput.Result.FlaggedResources, nil
}

func resourceToMetrics(resource *support.TrustedAdvisorResourceDetail) (prometheus.Metric, prometheus.Metric, error) {
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
		serviceLimit, prometheus.GaugeValue, float64(limitInt), *region, *service, *limitName,
	)
	usageMetric := prometheus.MustNewConstMetric(
		serviceUsage, prometheus.GaugeValue, float64(usageInt), *region, *service, *limitName,
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
