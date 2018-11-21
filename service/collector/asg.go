package collector

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

const (
	// ASGLabel is the metric's label key that will hold the ASG name.
	ASGLabel = "asg"

	// Subsystem will become the second part of the metric name, right after namespace.
	Subsystem = "asg"
)

var (
	asgDesiredDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "desired_count"),
		"Gauge about the number of EC2 instances that should be in the ASG.",
		[]string{
			ASGLabel,
			AccountLabel,
			ClusterLabel,
			InstallationLabel,
			OrganizationLabel,
		},
		nil,
	)

	asgInserviceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "inservice_count"),
		"Gauge about the number of EC2 instances in the ASG that are in state InService.",
		[]string{
			ASGLabel,
			AccountLabel,
			ClusterLabel,
			InstallationLabel,
			OrganizationLabel,
		},
		nil,
	)
)

// ASGConfig is this collector's configuration struct.
type ASGConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

// ASG is the main struct for this collector.
type ASG struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

// NewASG creates a new AutoScalingGroup metrics collector.
func NewASG(config ASGConfig) (*ASG, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	a := &ASG{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return a, nil
}

// Collect is the main metrics collection function.
func (a *ASG) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := a.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := a.collectForAccount(ch, awsClients)
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
func (a *ASG) Describe(ch chan<- *prometheus.Desc) error {
	ch <- asgDesiredDesc
	ch <- asgInserviceDesc
	return nil
}

// collectForAccount collects and emits metrics for one AWS account.
func (a *ASG) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	account, err := a.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	var autoScalingGroups []*autoscaling.Group
	{
		i := &autoscaling.DescribeAutoScalingGroupsInput{}
		o, err := awsClients.AutoScaling.DescribeAutoScalingGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}
		autoScalingGroups = o.AutoScalingGroups
	}

	for _, asg := range autoScalingGroups {
		var cluster, installation, organization string

		for _, tag := range asg.Tags {
			switch *tag.Key {
			case ClusterTag:
				cluster = *tag.Value
			case InstallationTag:
				installation = *tag.Value
			case OrganizationTag:
				organization = *tag.Value
			}
		}

		if installation != a.installationName {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			asgDesiredDesc,
			prometheus.GaugeValue,
			float64(*asg.DesiredCapacity),
			*asg.AutoScalingGroupName,
			account,
			cluster,
			installation,
			organization,
		)

		ch <- prometheus.MustNewConstMetric(
			asgInserviceDesc,
			prometheus.GaugeValue,
			float64(len(asg.Instances)),
			*asg.AutoScalingGroupName,
			account,
			cluster,
			installation,
			organization,
		)
	}

	return nil
}
