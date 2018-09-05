package adapter

import "github.com/giantswarm/aws-operator/service/controller/v16/key"

func asgType(config Config) string {
	return prefixWorker
}

func baseDomain(config Config) string {
	return key.BaseDomain(config.CustomObject)
}

func clusterID(config Config) string {
	return key.ClusterID(config.CustomObject)
}

func hostedZoneNameServers(config Config) string {
	return config.StackState.HostedZoneNameServers
}

func masterInstanceResourceName(config Config) string {
	return config.StackState.MasterInstanceResourceName
}

func route53Enabled(config Config) bool {
	return config.Route53Enabled
}

func workerImageID(config Config) string {
	return config.StackState.WorkerImageID
}
