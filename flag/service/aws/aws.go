package aws

import (
	"github.com/giantswarm/aws-operator/flag/service/aws/accesskey"
	"github.com/giantswarm/aws-operator/flag/service/aws/loggingbucket"
	"github.com/giantswarm/aws-operator/flag/service/aws/route53"
)

type AWS struct {
	AccessKey              accesskey.AccessKey
	AdvancedMonitoringEC2  string
	LoggingBucket          loggingbucket.LoggingBucket
	Encrypter              string
	HostAccessKey          accesskey.AccessKey
	IncludeTags            string
	PodInfraContainerImage string
	PubKeyFile             string
	Region                 string
	Route53                route53.Route53
	S3AccessLogsExpiration string
	VaultAddress           string
}
