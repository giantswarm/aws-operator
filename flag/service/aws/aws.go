package aws

import (
	"github.com/giantswarm/aws-operator/flag/service/aws/accesskey"
	"github.com/giantswarm/aws-operator/flag/service/aws/loggingbucket"
)

type AWS struct {
	AccessKey              accesskey.AccessKey
	AdvancedMonitoringEC2  string
	LoggingBucket          loggingbucket.LoggingBucket
	HostAccessKey          accesskey.AccessKey
	PubKeyFile             string
	Region                 string
	S3AccessLogsExpiration string
}
