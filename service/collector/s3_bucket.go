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
	// labelS3Bucket is the metric's label key that will hold the S3 bucket name.
	labelS3Bucket = "bucket_name"
)

const (
	// subsystemS3Bucket will become the second part of the metric name, right
	// after namespace.
	subsystemS3Bucket = "s3_bucket"
)

var (
	s3ObjectsTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemS3Bucket, "s3_objects_total"),
		"Gauge about the number of S3 objects within a S3 bucket.",
		[]string{
			labelS3Bucket,
			labelCluster,
		},
		nil,
	)
)

// S3BucketConfig is this collector's configuration struct.
type S3BucketConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

// S3Bucket is the main struct for this collector.
type S3Bucket struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

// NewS3Bucket creates a new AutoScalingGroup metrics collector.
func NewS3Bucket(config S3BucketConfig) (*S3Bucket, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	s := &S3Bucket{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return s, nil
}

// Collect is the main metrics collection function.
func (s *S3Bucket) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := s.helper.GetAWSClients()
	if err != nil {
		return microerror.Mask(err)
	}

	var g errgroup.Group

	for _, item := range awsClientsList {
		awsClients := item

		g.Go(func() error {
			err := s.collectForAccount(ch, awsClients)
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
func (s *S3Bucket) Describe(ch chan<- *prometheus.Desc) error {
	ch <- s3ObjectsTotalDesc
	return nil
}

// collectForAccount collects and emits metrics for one AWS account.
func (s *S3Bucket) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	account, err := s.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	var nextToken *string
	for {
		var autoScalingGroups []*autoscaling.Group
		{
			i := &autoscaling.DescribeAutoScalingGroupsInput{
				NextToken: nextToken,
			}
			o, err := awsClients.AutoScaling.DescribeAutoScalingGroups(i)
			if err != nil {
				return microerror.Mask(err)
			}
			autoScalingGroups = o.AutoScalingGroups
			nextToken = o.NextToken
		}

		for _, asg := range autoScalingGroups {
			var cluster, installation, organization string

			for _, tag := range asg.Tags {
				switch *tag.Key {
				case tagCluster:
					cluster = *tag.Value
				case tagInstallation:
					installation = *tag.Value
				case tagOrganization:
					organization = *tag.Value
				}
			}

			if installation != s.installationName {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				s3ObjectsTotalDesc,
				prometheus.GaugeValue,
				float64(*asg.DesiredCapacity),
				*asg.AutoScalingGroupName,
				account,
				cluster,
				installation,
				organization,
			)
		}

		if nextToken == nil {
			break
		}
	}

	return nil
}
