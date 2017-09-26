package host

import "github.com/giantswarm/aws-operator/flag/service/aws/host/accesskey"

type Host struct {
	AccessKey accesskey.AccessKey
	Region    string
}
