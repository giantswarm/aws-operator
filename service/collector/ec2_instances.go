package collector

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	// labelAvailabilityZone will contain the AZ name the instance is in
	labelAvailabilityZone = "availability_zone"

	// labelInstance is the metric's label key that will hold the ec2 instance ID.
	labelInstance = "ec2_instance"

	// labelInstanceState is a label that will contain the instance state string
	labelInstanceState = "state"

	// labelInstanceStatus is a label that will contain the instance status check value
	labelInstanceStatus = "status"

	// labelInstanceType will contain the instance type name
	labelInstanceType = "instance_type"

	// labelPrivateDNS will contain the private dns name
	labelPrivateDNS = "private_dns"

	// subsystemEC2 will become the second part of the metric name, right after namespace.
	subsystemEC2 = "ec2"
)

var (
	ec2InstanceStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystemEC2, "instance_status"),
		"Gauge indicating the status of an EC2 instance. 1 = healthy, 0 = unhealthy",
		[]string{
			labelInstance,
			labelAccount,
			labelCluster,
			labelInstallation,
			labelOrganization,
			labelAvailabilityZone,
			labelInstanceType,
			labelPrivateDNS,
			labelInstanceState,
			labelInstanceStatus,
		},
		nil,
	)
)

// EC2InstancesConfig is this collector's configuration struct.
type EC2InstancesConfig struct {
	Helper *helper
	Logger micrologger.Logger

	InstallationName string
}

// EC2Instances is the main struct for this collector.
type EC2Instances struct {
	helper *helper
	logger micrologger.Logger

	installationName string
}

// NewEC2Instances creates a new EC2 instance metrics collector.
func NewEC2Instances(config EC2InstancesConfig) (*EC2Instances, error) {
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	e := &EC2Instances{
		helper: config.Helper,
		logger: config.Logger,

		installationName: config.InstallationName,
	}

	return e, nil
}

// Collect is the main metrics collection function.
func (e *EC2Instances) Collect(ch chan<- prometheus.Metric) error {
	awsClientsList, err := e.helper.GetAWSClients()
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

// Describe emits the description for the metrics collected here.
func (e *EC2Instances) Describe(ch chan<- *prometheus.Desc) error {
	ch <- ec2InstanceStatus
	return nil
}

// collectForAccount collects and emits our metric for one AWS account.
//
// We gather two separate collections first, then match them by instance ID:
// - instance information, including tags, only for those tagged for our installation
// - instance status information
func (e *EC2Instances) collectForAccount(ch chan<- prometheus.Metric, awsClients clientaws.Clients) error {
	account, err := e.helper.AWSAccountID(awsClients)
	if err != nil {
		return microerror.Mask(err)
	}

	// Collect instance status info.
	// map key will be the instance ID.
	instanceStatuses := map[string]*ec2.InstanceStatus{}
	{
		input := &ec2.DescribeInstanceStatusInput{
			IncludeAllInstances: aws.Bool(true),
			MaxResults:          aws.Int64(1000),
		}

		for {
			o, err := awsClients.EC2.DescribeInstanceStatus(input)
			if err != nil {
				return microerror.Mask(err)
			}

			// collect statuses
			for _, s := range o.InstanceStatuses {
				instanceStatuses[*s.InstanceId] = s
			}

			if o.NextToken == nil {
				break
			}
			input.SetNextToken(*o.NextToken)
		}
	}

	// Collect instance info.
	instances := map[string]*ec2.Instance{}
	{
		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagInstallation)),
					Values: []*string{
						aws.String(e.installationName),
					},
				},
			},
			MaxResults: aws.Int64(1000),
		}

		for {
			o, err := awsClients.EC2.DescribeInstances(input)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, reservation := range o.Reservations {
				for _, instance := range reservation.Instances {
					instances[*instance.InstanceId] = instance
				}
			}

			if o.NextToken == nil {
				break
			}
			input.SetNextToken(*o.NextToken)
		}
	}

	// Iterate over found instances and emit metrics.
	for instanceID := range instances {
		// Skip if we don't have a status for this instance.
		statuses, statusesAvailable := instanceStatuses[instanceID]
		if !statusesAvailable {
			continue
		}

		var az, cluster, instanceType, installation, organization, privateDNS, state, status string
		for _, tag := range instances[instanceID].Tags {
			switch *tag.Key {
			case tagCluster:
				cluster = *tag.Value
			case key.TagInstallation:
				installation = *tag.Value
			case tagOrganization:
				organization = *tag.Value
			}
		}

		instanceType = *instances[instanceID].InstanceType
		privateDNS = *instances[instanceID].PrivateDnsName

		up := 0
		if statuses.InstanceState.Name != nil {
			state = strings.ToLower(*statuses.InstanceState.Name)
		}
		if statuses.InstanceStatus.Status != nil {
			status = strings.ToLower(*statuses.InstanceStatus.Status)
		}
		if statuses.AvailabilityZone != nil {
			az = strings.ToLower(*statuses.AvailabilityZone)
		}
		if state == "running" && status == "ok" {
			up = 1
		}

		ch <- prometheus.MustNewConstMetric(
			ec2InstanceStatus,
			prometheus.GaugeValue,
			float64(up),
			instanceID,
			account,
			cluster,
			installation,
			organization,
			az,
			instanceType,
			privateDNS,
			state,
			status,
		)
	}

	return nil
}
