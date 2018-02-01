package alerter

import (
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
)

const (
	alertIntervalMins = 5
	vpcResourceType   = "vpc"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	// Settings.
	AwsConfig        awsutil.Config
	InstallationName string
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		G8sClient: nil,
		Logger:    nil,

		// Settings.
		AwsConfig:        awsutil.Config{},
		InstallationName: "",
	}
}

// Service implements the service interface.
type Service struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	installationName string

	// Internals.
	awsClients awsutil.Clients
	g8sClient  versioned.Interface
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.G8sClient must not be empty")
	}

	// Settings.
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.installationName must not be empty")
	}

	// Internals.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}

	awsClients := awsutil.NewClients(config.AwsConfig)

	newService := &Service{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		installationName: config.InstallationName,

		// Internals.
		awsClients: awsClients,
		g8sClient:  config.G8sClient,
	}

	return newService, nil
}

// StartAlerts starts a background ticker that checks for orphan resources.
func (s *Service) StartAlerts() {
	go func() {
		s.logger.Log("info", "starting alerter service to check for orphan resources")

		alertChan := time.NewTicker(time.Minute * alertIntervalMins).C

		for {
			select {
			case <-alertChan:
				err := s.RunAllChecks()
				if err != nil {
					s.logger.Log("error", fmt.Sprintf("could not execute run all checks: '%#v'", err))
				}
			}
		}
	}()
}

// RunAllChecks looks for problems with clusters that we want to alert on.
func (s *Service) RunAllChecks() error {
	clusterIDs, err := s.ListClusters()
	if err != nil {
		return microerror.Mask(err)
	}


	err = s.OrphanResourcesAlert(clusterIDs)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
// OrphanResourcesAlert looks for AWS resources not associated with a cluster.
func (s *Service) OrphanResourcesAlert(clusterIDs []string) error {
	vpcNames, err := s.ListVpcs()
	if err != nil {
		return microerror.Mask(err)
	}

	orphanVpcs := FindOrphanResources(clusterIDs, vpcNames)
	s.UpdateOrphanResourceMetrics(vpcResourceType, orphanVpcs)

	return nil
}

// ListClusters lists the cluster custom objects.
func (s Service) ListClusters() ([]string, error) {
	clusterIDs := []string{}

	awsConfigs, err := s.g8sClient.ProviderV1alpha1().AWSConfigs("").List(v1.ListOptions{})
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	for _, awsConfig := range awsConfigs.Items {
		clusterIDs = append(clusterIDs, key.ClusterID(awsConfig))
	}

	return clusterIDs, nil
}

// FindOrphanResources compares a list of cluster IDs and resource names. It
// returns any resources not associated with a cluster.
func FindOrphanResources(clusterIDs []string, resourceNames []string) []string {
	clusters := map[string]bool{}
	orphanResources := []string{}

	for _, clusterID := range clusterIDs {
		clusters[clusterID] = true
	}

	for _, resourceName := range resourceNames {
		if ok, _ := clusters[resourceName]; !ok {
			orphanResources = append(orphanResources, resourceName)
		}
	}

	return orphanResources
}

// UpdateOrphanResourceMetrics updates the metric and logs the results.
func (s Service) UpdateOrphanResourceMetrics(resourceType string, resourceNames []string) {
	resourceCount := len(resourceNames)

	orphanResourcesTotal.WithLabelValues(resourceType).Set(float64(resourceCount))
	s.logger.Log("info", fmt.Sprintf("alerter service found %d %s resources not associated with a cluster", resourceCount, resourceType))

	if resourceCount > 0 {
		s.logger.Log("info", fmt.Sprintf("orphan %s names are %s", resourceType, strings.Join(resourceNames, ",")))
	}
}
