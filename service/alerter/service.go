package alerter

import (
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v2/key"
)

const (
	alertIntervalMins      = 5
	masterNodeResourceType = "master_node"
	vpcResourceType        = "vpc"
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
	awsConfigs, err := s.ListClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	clusterIDs, err := s.ListClusterIDs(awsConfigs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.FindDuplicateMasterNodes(clusterIDs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.OrphanResourcesAlert(clusterIDs)
	if err != nil {
		return microerror.Mask(err)
	}

	err = s.OrphanClustersAlert(awsConfigs)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// FindDuplicateMasterNodes looks for clusters with duplicate master nodes
// which is an error state.
func (s *Service) FindDuplicateMasterNodes(clusterIDs []string) error {
	affectedClusters := []string{}

	for _, clusterID := range clusterIDs {
		masterInstances, err := s.ListMasterInstances(clusterID)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(masterInstances) > 1 {
			affectedClusters = append(affectedClusters, clusterID)
		}
	}

	s.UpdateDuplicateResourceMetrics(masterNodeResourceType, affectedClusters)

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

// OrphanClustersAlert looks for clusters without AWS resources associated.
func (s *Service) OrphanClustersAlert(awsConfigs []v1alpha1.AWSConfig) error {
	vpcNames, err := s.ListVpcs()
	if err != nil {
		return microerror.Mask(err)
	}

	orphanClusters := FindOrphanClusters(awsConfigs, vpcNames)
	s.UpdateOrphanClusterMetrics(vpcResourceType, orphanClusters)

	return nil
}

// ListClusters lists the cluster custom objects.
func (s Service) ListClusters() ([]v1alpha1.AWSConfig, error) {
	awsConfigs, err := s.g8sClient.ProviderV1alpha1().AWSConfigs("").List(v1.ListOptions{})
	if err != nil {
		return []v1alpha1.AWSConfig{}, microerror.Mask(err)
	}

	return awsConfigs.Items, nil
}

// ListClusterIDs lists the cluster custom objects IDs.
func (s Service) ListClusterIDs(awsConfigs []v1alpha1.AWSConfig) ([]string, error) {
	clusterIDs := []string{}

	for _, awsConfig := range awsConfigs {
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

// FindOrphanClusters compares a list of cluster IDs and resource names. It
// returns the IDs of any cluster created a reasonable amount of time ago without
// resources associated.
func FindOrphanClusters(awsConfigs []v1alpha1.AWSConfig, resourceNames []string) []string {
	resources := map[string]bool{}
	orphanClusters := []string{}

	for _, resourceName := range resourceNames {
		resources[resourceName] = true
	}

	n := time.Now()
	l := n.Add(-10 * time.Minute)
	for _, awsConfig := range awsConfigs {
		// if the cluster has been recently created maybe the resources are not still there.
		if awsConfig.CreationTimestamp.After(l) {
			continue
		}
		clusterID := key.ClusterID(awsConfig)
		if ok := resources[clusterID]; !ok {
			orphanClusters = append(orphanClusters, clusterID)
		}
	}

	return orphanClusters
}

// UpdateDuplicateResourceMetrics updates the metric and logs the results.
func (s Service) UpdateDuplicateResourceMetrics(resourceType string, clusterIDs []string) {
	resourceCount := len(clusterIDs)

	duplicateResourcesTotal.WithLabelValues(resourceType).Set(float64(resourceCount))
	s.logger.Log("info", fmt.Sprintf("alerter service found %d clusters with duplicate resources", resourceCount))

	if resourceCount > 0 {
		s.logger.Log("info", fmt.Sprintf("clusters with duplicate %s %s", resourceType, strings.Join(clusterIDs, ",")))
	}
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

// UpdateOrphanClusterMetrics updates the metric and logs the results.
func (s Service) UpdateOrphanClusterMetrics(resourceType string, clusterIDs []string) {
	resourceCount := len(clusterIDs)

	orphanClustersTotal.WithLabelValues(resourceType).Set(float64(resourceCount))

	s.logger.Log("info", fmt.Sprintf("alerter service found %d clusters with missing resources", resourceCount))

	if resourceCount > 0 {
		s.logger.Log("info", fmt.Sprintf("clusters with missing %s %s", resourceType, strings.Join(clusterIDs, ",")))
	}
}
