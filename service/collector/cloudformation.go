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
	labelCloudFormation = "cloudformation"

	subsystemCloudFormation = "cloudformation"
)

var (
	cloudFormationStackDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemCloudFormation, "info"),
		"CloudFormation information.",
		[]string{
			labelAccountID,
			labelCluster,
			labelID,
			labelInstallation,
			labelName,
			labelOrganization,
			labelStack,
			labelState,
		},
		nil,
	)
)

type CloudFormationConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type CloudFormation struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

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

func (cf *CloudFormation) Describe(ch chan<- *prometheus.Desc) error {
	ch <- cloudFormationStackDesc
	return nil
}

func (cf *CloudFormation) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	o, err := awsClients.CloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	accountID, err := cf.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, stack := range o.Stacks {
		var cluster, installation, name, organization, stackName string

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
				stackName = *tag.Value
			}
		}

		if installation != cf.installationName {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			cloudFormationStackDesc,
			prometheus.GaugeValue,
			GaugeValue,
			accountID,
			cluster,
			*stack.StackId,
			installation,
			name,
			organization,
			stackName,
			*stack.StackStatus,
		)
	}

	return nil
}
