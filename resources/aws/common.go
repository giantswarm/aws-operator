package aws

import (
	awsutil "github.com/giantswarm/aws-operator/client/aws"
)

type AWSEntity struct {
	Clients     awsutil.Clients
	HostClients awsutil.Clients
}
