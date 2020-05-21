package aws

import (
	"github.com/giantswarm/aws-operator/flag/service/aws/accesskey"
	"github.com/giantswarm/aws-operator/flag/service/aws/cni"
	"github.com/giantswarm/aws-operator/flag/service/aws/loggingbucket"
	"github.com/giantswarm/aws-operator/flag/service/aws/route53"
	"github.com/giantswarm/aws-operator/flag/service/aws/trustedadvisor"
)

type AWS struct {
	AccessKey              accesskey.AccessKey
	AdvancedMonitoringEC2  string
	AvailabilityZones      string
	HostAccessKey          accesskey.AccessKey
	IncludeTags            string
	LoggingBucket          loggingbucket.LoggingBucket
	PodInfraContainerImage string
	Region                 string
	Route53                route53.Route53
	RouteTables            string
	S3AccessLogsExpiration string
	TrustedAdvisor         trustedadvisor.TrustedAdvisor
	VaultAddress           string
	CNI                    cni.CNI
}
