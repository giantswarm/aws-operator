package collector

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	NameTag      = "Name"
	StackNameTag = "aws:cloudformation:stack-name"

	AccountIdLabel = "account_id"
	CidrLabel      = "cidr"
	IDLabel        = "id"
	NameLabel      = "name"
	StackNameLabel = "stack_name"
	StateLabel     = "state"
)

var (
	vpcsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "vpc_info"),
		"VPC information.",
		[]string{
			AccountIdLabel,
			CidrLabel,
			ClusterLabel,
			IDLabel,
			InstallationLabel,
			NameLabel,
			OrganizationLabel,
			StackNameLabel,
			StateLabel,
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
	awsClientsList, err := v.helper.GetAWSClients()
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
			case ClusterTag:
				cluster = *tag.Value
			case InstallationTag:
				installation = *tag.Value
			case NameTag:
				name = *tag.Value
			case OrganizationTag:
				organization = *tag.Value
			case StackNameTag:
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
