package aws

import (
	"github.com/giantswarm/aws-operator/flag/service/aws/accesskey"
)

type AWS struct {
	AccessKey              accesskey.AccessKey
	AdvancedMonitoringEC2  string
	HostAccessKey          accesskey.AccessKey
	PubKeyFile             string
	Region                 string
	S3AccessLogsExpiration string
}
