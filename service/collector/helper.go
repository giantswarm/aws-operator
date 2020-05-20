package collector

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	clientaws "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/pkg/project"
	"github.com/giantswarm/aws-operator/service/internal/accountid"
	"github.com/giantswarm/aws-operator/service/internal/credential"
)

type helperConfig struct {
	Clients k8sclient.Interface
	Logger  micrologger.Logger

	AWSConfig clientaws.Config
}

type helper struct {
	clients k8sclient.Interface
	logger  micrologger.Logger

	awsConfig clientaws.Config
}

func newHelper(config helperConfig) (*helper, error) {
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var emptyAWSConfig clientaws.Config
	if config.AWSConfig == emptyAWSConfig {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSConfig must not be empty", config)
	}

	h := &helper{
		clients: config.Clients,
		logger:  config.Logger,

		awsConfig: config.AWSConfig,
	}

	return h, nil
}

// GetARNs list all unique aws IAM ARN from credential secret.
func (h *helper) GetARNs(clusterList *infrastructurev1alpha2.AWSClusterList) ([]string, error) {
	var arns []string

	// Get unique ARNs.
	arnsMap := make(map[string]bool)
	for _, clusterCR := range clusterList.Items {
		arn, err := credential.GetARN(h.clients.K8sClient(), clusterCR)
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
	arn, err := credential.GetDefaultARN(h.clients.K8sClient())
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
func (h *helper) GetAWSClients(clusterList *infrastructurev1alpha2.AWSClusterList) ([]clientaws.Clients, error) {
	var (
		clients    []clientaws.Clients
		clientsMap = make(map[string]clientaws.Clients)
	)

	arns, err := h.GetARNs(clusterList)
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

// ListReconciledClusters provides a list of clusters reconciled by this
// particular operator version.
func (h *helper) ListReconciledClusters() (*infrastructurev1alpha2.AWSClusterList, error) {
	ctx := context.Background()

	clusters := &infrastructurev1alpha2.AWSClusterList{}
	err := h.clients.CtrlClient().List(
		ctx,
		clusters,
		runtimeclient.MatchingLabels{
			label.OperatorVersion: project.Version(),
		},
	)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return clusters, err
}
