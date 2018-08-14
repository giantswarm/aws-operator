package collector

import (
	"fmt"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/controller/v15/credential"
)

// getARNs list all unique aws IAM ARN from credential secret.
func (c Collector) getARNs() ([]string, error) {
	var arns []string

	// List AWSConfigs.
	awsConfigClient := c.g8sClient.ProviderV1alpha1().AWSConfigs("")
	awsConfigs, err := awsConfigClient.List(v1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Get unique ARNs.
	arnsMap := make(map[string]bool)
	for _, awsConfig := range awsConfigs.Items {
		arn, err := credential.GetARN(c.k8sClient, &awsConfig)
		// Collect as many ARNs as possible in order to provide most metrics.
		// Ignore old cluster which do not have credential.
		if credential.IsCredentialNameEmptyError(err) ||
			credential.IsCredentialNamespaceEmptyError(err) {
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		arnsMap[arn] = true
	}

	// Ensure we check the default guest account for old cluster not having credential.
	arn, err := credential.GetDefaultARN(c.k8sClient)
	if err == nil {
		arnsMap[arn] = true
	}

	for arn, _ := range arnsMap {
		arns = append(arns, arn)
	}

	return arns, nil
}

// getAWSClients return a list of aws clients for every guest cluster account plus
// the host cluster account.
func (c Collector) getAWSClients() ([]aws.Clients, error) {
	var (
		clients    []aws.Clients
		clientsMap = make(map[string]aws.Clients)
	)

	arns, err := c.getARNs()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// addClientFunc add awsClients to clients using account id as key to guaranatee uniqueness.
	addClientFunc := func(awsClients aws.Clients, clients *map[string]aws.Clients) error {
		accountID, err := c.awsAccountID(awsClients)
		if err != nil {
			return microerror.Mask(err)
		}

		_, ok := (*clients)[accountID]
		if !ok {
			(*clients)[accountID] = awsClients
		}

		return nil
	}

	// Host cluster account.
	awsClients := aws.NewClients(c.awsConfig)
	err = addClientFunc(awsClients, &clientsMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Guest cluster accounts.
	for _, arn := range arns {
		awsConfig := c.awsConfig
		awsConfig.RoleARN = arn

		awsClients := aws.NewClients(awsConfig)
		err = addClientFunc(awsClients, &clientsMap)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// Convert map to slice.
	for accountID, client := range clientsMap {
		clients = append(clients, client)
		c.logger.Log("level", "debug", "message", fmt.Sprintf("collecting metrics in account: %s", accountID))
	}

	return clients, nil
}

// awsAccountID return the AWS account ID.
func (c Collector) awsAccountID(awsClients aws.Clients) (string, error) {
	config := awsservice.Config{
		Clients: awsservice.Clients{
			KMS: awsClients.KMS,
			STS: awsClients.STS,
		},
		Logger: c.logger,
	}

	awsService, err := awsservice.New(config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	accountID, err := awsService.GetAccountID()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return accountID, nil
}
