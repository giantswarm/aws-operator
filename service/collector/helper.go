package collector

import (
	"fmt"
	"time"

	v1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/internal/accountid"
	"github.com/giantswarm/aws-operator/service/internal/credential"
)

type helperConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	AWSConfig clientaws.Config
}

type helper struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	g8sCache cache.Store

	awsConfig clientaws.Config
}

func newHelper(config helperConfig) (*helper, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var emptyAWSConfig clientaws.Config
	if config.AWSConfig == emptyAWSConfig {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSConfig must not be empty", config)
	}

	var store cache.Store
	{
		listerWatcher := &ListerWatcher{
			clusters: config.G8sClient.InfrastructureV1alpha2().AWSClusters(metav1.NamespaceAll),
		}
		store = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
		reflector := cache.NewReflector(listerWatcher, &v1alpha2.AWSCluster{}, store, 2*time.Minute)
		go reflector.Run(make(<-chan struct{}))
		// force 1st reflector sync
		listerWatcher.List(metav1.ListOptions{})
	}

	h := &helper{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		g8sCache: store,

		awsConfig: config.AWSConfig,
	}

	return h, nil
}

// GetARNs list all unique aws IAM ARN from credential secret.
func (h *helper) GetARNs() ([]string, error) {
	var arns []string

	clusterCRList, err := h.g8sClient.InfrastructureV1alpha2().AWSClusters(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get unique ARNs.
	arnsMap := make(map[string]bool)
	for _, clusterCR := range clusterCRList.Items {
		arn, err := credential.GetARN(h.k8sClient, clusterCR)
		// Collect as many ARNs as possible in order to provide most metrics.
		// Ignore old cluster which do not have credential.
		if credential.IsCredentialNameEmptyError(err) {
			continue
		} else if credential.IsCredentialNamespaceEmptyError(err) {
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		arnsMap[arn] = true
	}

	// Ensure we check the default guest account for old cluster not having credential.
	arn, err := credential.GetDefaultARN(h.k8sClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	arnsMap[arn] = true

	for arn := range arnsMap {
		arns = append(arns, arn)
	}

	return arns, nil
}

// GetAWSClients return a list of aws clients for every guest cluster account plus
// the host cluster account.
func (h *helper) GetAWSClients() ([]clientaws.Clients, error) {
	var (
		clients    []clientaws.Clients
		clientsMap = make(map[string]clientaws.Clients)
	)

	arns, err := h.GetARNs()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// addClientFunc add awsClients to clients using account id as key to guaranatee uniqueness.
	addClientFunc := func(awsClients clientaws.Clients, clients map[string]clientaws.Clients) error {
		accountID, err := h.AWSAccountID(awsClients)
		if err != nil {
			return microerror.Mask(err)
		}

		_, ok := clients[accountID]
		if !ok {
			clients[accountID] = awsClients
		}

		return nil
	}

	// Control plane account.
	awsClients, err := clientaws.NewClients(h.awsConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	err = addClientFunc(awsClients, clientsMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Tenant cluster accounts.
	for _, arn := range arns {
		awsConfig := h.awsConfig
		awsConfig.RoleARN = arn

		awsClients, err := clientaws.NewClients(awsConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		err = addClientFunc(awsClients, clientsMap)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Convert map to slice.
	for accountID, client := range clientsMap {
		clients = append(clients, client)
		h.logger.Log("level", "debug", "message", fmt.Sprintf("collecting metrics in account: %s", accountID))
	}

	return clients, nil
}

// AWSAccountID return the AWS account ID.
func (h *helper) AWSAccountID(awsClients clientaws.Clients) (string, error) {
	var err error

	var accountIDService *accountid.AccountID
	{
		c := accountid.Config{
			Logger: h.logger,
			STS:    awsClients.STS,
		}

		accountIDService, err = accountid.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	accountID, err := accountIDService.Lookup()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return accountID, nil
}

// IsClusterReconciledByThisVersion checks whether an AWSCluster object with
// the given name is reconciled by this particular version of aws-operator
// (based on version found in pkg/project/project.go).
func (h *helper) IsClusterReconciledByThisVersion(clusterName string) (bool, error) {
	if clusterName == "" {
		return false, microerror.Maskf(executionFailedError, "empty cluster name")
	}
	key := fmt.Sprintf("default/%s", clusterName)

	logger.Errorf("KUBA helper: searching %q", key)
	item, exists, err := h.g8sCache.GetByKey(key)
	logger.Errorf("KUBA helper: item:%+v, exists:%v, err:%v", item, exists, err)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if !exists {
		return false, microerror.Maskf(executionFailedError, "could not find key in cache: %#q", key)
	}

	m, err := meta.Accessor(item)
	if err != nil {
		return false, microerror.Mask(err)
	}
	logger.Errorf("KUBA helper: got accessor; labels: %+v", m.GetLabels())

	// This label contains version of operator reconciling this particular
	// resource. For more information see:
	// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md
	versionAssigned, ok := m.GetLabels()["aws-operator.giantswarm.io/version"]
	if !ok {
		return false, microerror.Maskf(
			executionFailedError,
			"could not find aws-operator.giantswarm.io/version for AWSCluster %q/%q",
			m.GetNamespace(), m.GetName(),
		)
	}

	logger.Errorf("KUBA helper: comparing cluster (%v) to project (%v)", versionAssigned, project.Version())
	if versionAssigned == project.Version() {
		return true, nil
	}
	return false, nil
}
