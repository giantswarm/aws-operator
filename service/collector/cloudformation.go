package collector

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

const (
	labelStackType   = "stack_type"
	labelAccountType = "account_type"
)

const (
	// Second part of the metric name, right after namespace.
	subsystemCloudFormation = "cloudformation"
)

var (
	cloudFormationStackDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCloudFormation, "info"),
		"CloudFormation information.",
		[]string{
			labelAccountID,
			labelAccountType,
			labelCluster,
			labelID,
			labelInstallation,
			labelName,
			labelOrganization,
			labelStackType,
			labelState,
		},
		nil,
	)
)

// Configuration struct.
type CloudFormationConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

// Main struct for this collector.
type CloudFormation struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

// Creates a new CloudFormation metrics collector.
func NewCloudFormation(config CloudFormationConfig) (*CloudFormation, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	cf := &CloudFormation{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return cf, nil
}

// Collect is the main metrics collection function.
func (cf *CloudFormation) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := cf.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := cf.collectForAccount(ch, awsClients)
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

// Describe emits the description for the metrics collected here.
func (cf *CloudFormation) Describe(ch chan<- *prometheus.Desc) error {
	ch <- cloudFormationStackDesc
	return nil
}

// collectForAccount collects metrics for one AWS account.
func (cf *CloudFormation) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	o, err := awsClients.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	accountID, err := cf.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}
	accountType := "" // TODO: specify if its tenant-cluster or control plane account

	for _, stack := range o.Stacks {
		var cluster, installation, name, organization, stackType string

		for _, tag := range stack.Tags {
			switch *tag.Key {
			case tagCluster:
				cluster = *tag.Value
			case tagInstallation:
				installation = *tag.Value
			case tagName:
				name = *tag.Value
			case tagOrganization:
				organization = *tag.Value
			case tagStack:
				stackType = *tag.Value
			}
		}

		if installation != cf.installationName {
			continue
		}

		if !isOwnStack(stackType) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			cloudFormationStackDesc,
			prometheus.GaugeValue,
			GaugeValue,
			accountID,
			accountType,
			cluster,
			*stack.StackId,
			installation,
			name,
			organization,
			stackType,
			*stack.StackStatus,
		)
	}

	return nil
}

// Check if the input stack is our own by checking the name of the stack type
func isOwnStack(StackType string) bool {
	return StackType == "tccp" || StackType == "tccpf" || StackType == "tccpi" || StackType == "tcnp" || StackType == "tcnpf"
}
