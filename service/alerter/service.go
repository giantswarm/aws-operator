package alerter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/tpr"
	"k8s.io/client-go/kubernetes"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/service/keyv1"
)

const (
	alertIntervalMins = 5
	vpcResourceType   = "vpc"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
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
		K8sClient: nil,
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
	k8sClient  kubernetes.Interface
	tpr        *tpr.TPR
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
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

	var err error

	var newTPR *tpr.TPR
	{
		tprConfig := tpr.DefaultConfig()

		tprConfig.K8sClient = config.K8sClient
		tprConfig.Logger = config.Logger

		tprConfig.Description = awstpr.Description
		tprConfig.Name = awstpr.Name
		tprConfig.Version = awstpr.VersionV1

		newTPR, err = tpr.New(tprConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		installationName: config.InstallationName,

		// Internals.
		awsClients: awsClients,
		k8sClient:  config.K8sClient,
		tpr:        newTPR,
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
				err := s.OrphanResourcesAlert()
				if err != nil {
					s.logger.Log("error", fmt.Sprintf("could not execute orphan resources alert: '%#v'", err))
				}
			}
		}
	}()
}

// OrphanResourcesAlert looks for AWS resources not associated with a cluster.
func (s *Service) OrphanResourcesAlert() error {
	clusterIDs, err := s.ListClusters()
	if err != nil {
		return microerror.Mask(err)
	}

	vpcNames, err := s.ListVpcs()
	if err != nil {
		return microerror.Mask(err)
	}

	orphanVpcs := FindOrphanResources(clusterIDs, vpcNames)
	s.UpdateMetrics(vpcResourceType, orphanVpcs)

	return nil
}

// ListClusters lists the cluster custom objects.
func (s Service) ListClusters() ([]string, error) {
	clusterIDs := []string{}

	endpoint := s.tpr.Endpoint("")
	restClient := s.k8sClient.Core().RESTClient()

	clientResponse := restClient.Get().AbsPath(endpoint).Do()
	b, err := clientResponse.Raw()
	if err != nil {
		return []string{}, microerror.Mask(err)
	}

	clusterList := &awstpr.List{}
	if err := json.Unmarshal(b, clusterList); err != nil {
		return []string{}, microerror.Mask(err)
	}

	for _, cluster := range clusterList.Items {
		clusterIDs = append(clusterIDs, keyv1.ClusterID(*cluster))
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

// UpdateMetrics updates the metric and logs the results.
func (s Service) UpdateMetrics(resourceType string, resourceNames []string) {
	resourceCount := len(resourceNames)

	orphanResourcesTotal.WithLabelValues(resourceType).Set(float64(resourceCount))
	s.logger.Log("info", fmt.Sprintf("alerter service found %d %s resources not associated with a cluster", resourceCount, resourceType))

	if resourceCount > 0 {
		s.logger.Log("info", fmt.Sprintf("orphan %s names are %s", resourceType, strings.Join(resourceNames, ",")))
	}
}
