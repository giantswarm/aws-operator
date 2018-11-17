package collector

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/support"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/giantswarm/aws-operator/client/aws"
)

const (
	// categoryServiceLimit is the category returned by Trusted Advisor for checks
	// related to service limits and usage.
	categoryServiceLimit = "service_limits"
)

const (
	indexRegion  = 0
	indexService = 1
	indexName    = 2
	indexLimit   = 3
	indexUsage   = 4
)

const (
	// resourceMetadataLength is the length of resource metadata we expect.
	resourceMetadataLength = 6
)

const (
	labelAccountID = "account_id"
	labelRegion    = "region"
	labelService   = "service"
	labelName      = "name"
)

var (
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
	serviceLimit *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "service_limit"),
		"Service limits as reported by Trusted Advisor.",
		[]string{
			labelAccountID,
			labelRegion,
			labelService,
			labelName,
		},
		nil,
	)
	serviceUsage *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "service_usage"),
		"Service usage as reported by Trusted Advisor.",
		[]string{
			labelAccountID,
			labelRegion,
			labelService,
			labelName,
		},
		nil,
	)
	trustedAdvisorSupport *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "trusted_advisor_supported"),
		"Gauge describing whether Trusted Advisor is available with your support plan.",
		nil, nil,
	)
)

type TrustedAdvisorConfig struct {
	Helper *helper
	Logger micrologger.Logger
}

type TrustedAdvisor struct {
	helper *helper
	logger micrologger.Logger
}

func NewTrustedAdvisor(config TrustedAdvisorConfig) (*TrustedAdvisor, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &TrustedAdvisor{
		helper: config.Helper,
		logger: config.Logger,
	}

	return t, nil
}

func (t *TrustedAdvisor) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := t.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := t.collectForAccount(ch, awsClients)
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

func (t *TrustedAdvisor) Describe(ch chan<- *prometheus.Desc) error {
	ch <- serviceLimit
	ch <- serviceUsage
	ch <- trustedAdvisorSupport
	return nil
}

func (t *TrustedAdvisor) collectForAccount(ch chan<- prometheus.Metric, awsClients aws.Clients) error {
	accountID, err := t.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	checks, err := t.getTrustedAdvisorChecks(awsClients)
	if IsUnsupportedPlan(err) {
		ch <- trustedAdvisorNotSupported()
		return microerror.Mask(err)
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		ch <- trustedAdvisorSupported()
	}

	var g errgroup.Group

	for _, check := range checks {
		// Ignore any checks that don't relate to service limits.
		if *check.Category != categoryServiceLimit {
			continue
		}

		// Register the check ID for the current loop scope so it can safely be used
		// in the goroutine below, which is execute in parallel.
		id := check.Id

		g.Go(func() error {
			resources, err := t.getTrustedAdvisorResources(id, awsClients)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, resource := range resources {
				// One Trusted Advisor check returns the nil string for current usage.
				// Skip it.
				if len(resource.Metadata) == 6 && resource.Metadata[4] == nil {
					continue
				}

				limit, usage, err := resourceToMetrics(resource, accountID)
				if err != nil {
					return microerror.Mask(err)
				}

				ch <- limit
				ch <- usage
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

// getTrustedAdvisorCheckDescriptions calls Trusted Advisor API to get all
// available checks.
func (t *TrustedAdvisor) getTrustedAdvisorChecks(awsClients aws.Clients) ([]*support.TrustedAdvisorCheckDescription, error) {
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

// getTrustedAdvisorResources calls Trusted Advisor API to get flagged resources
// of the given check ID.
func (t *TrustedAdvisor) getTrustedAdvisorResources(id *string, awsClients aws.Clients) ([]*support.TrustedAdvisorResourceDetail, error) {
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

	region := resource.Metadata[indexRegion]
	service := resource.Metadata[indexService]
	limitName := resource.Metadata[indexName]

	limit := resource.Metadata[indexLimit]
	usage := resource.Metadata[indexUsage]

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
