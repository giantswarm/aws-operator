package collector

import (
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
		if err != nil {
			return nil, microerror.Mask(err)
		}
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
	var clients []aws.Clients

	arns, err := c.getARNs()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Host cluster account.
	clients = append(clients, aws.NewClients(c.awsConfig))

	// Guest cluster accounts.
	for _, arn := range arns {
		awsConfig := c.awsConfig
		awsConfig.RoleARN = arn

		clients = append(clients, aws.NewClients(awsConfig))
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
