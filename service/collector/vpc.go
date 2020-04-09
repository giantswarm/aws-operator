package collector

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	labelCIDR  = "cidr"
	labelID    = "id"
	labelStack = "stack_name"
	labelState = "state"
)

const (
	subsystemVPC = "vpc"
)

var (
	vpcsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemVPC, "info"),
		"VPC information.",
		[]string{
			labelAccountID,
			labelCIDR,
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

type VPCConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

type VPC struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

func NewVPC(config VPCConfig) (*VPC, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	v := &VPC{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return v, nil
}

func (v *VPC) Collect(ch chan<- prometheus.Metric) error {
	reconciledClusters, err := v.helper.ListReconciledClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	awsClientsList, err := v.helper.GetAWSClients(reconciledClusters)
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := v.collectForAccount(ch, awsClients)
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

func (v *VPC) Describe(ch chan<- *prometheus.Desc) error {
	ch <- vpcsDesc
	return nil
}

func (v *VPC) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	o, err := awsClients.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return microerror.Mask(err)
	}

	accountID, err := v.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, vpc := range o.Vpcs {
		var cluster, installation, name, organization, stackName string

		for _, tag := range vpc.Tags {
			switch *tag.Key {
			case tagCluster:
				cluster = *tag.Value
			case key.TagInstallation:
				installation = *tag.Value
			case tagName:
				name = *tag.Value
			case tagOrganization:
				organization = *tag.Value
			case tagStackName:
				stackName = *tag.Value
			}
		}

		if installation != v.installationName {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			vpcsDesc,
			prometheus.GaugeValue,
			GaugeValue,
			accountID,
			*vpc.CidrBlock,
			cluster,
			*vpc.VpcId,
			installation,
			name,
			organization,
			stackName,
			*vpc.State,
		)
	}

	return nil
}
