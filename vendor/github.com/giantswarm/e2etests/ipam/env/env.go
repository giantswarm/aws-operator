package env

import (
	"fmt"
	"os"
)

const (
	EnvVarAWSAPIHostedZoneGuest     = "AWS_API_HOSTED_ZONE_GUEST"
	EnvVarAWSIngressHostedZoneGuest = "AWS_INGRESS_HOSTED_ZONE_GUEST"
	EnvVarAWSRegion                 = "AWS_REGION"
	EnvVarAWSRouteTable0            = "AWS_ROUTE_TABLE_0"
	EnvVarAWSRouteTable1            = "AWS_ROUTE_TABLE_1"
	EnvVarClusterName               = "CLUSTER_NAME"
	EnvVarCommonDomain              = "COMMON_DOMAIN"
	EnvVarSSHPublickey              = "IDRSA_PUB"
	EnvVarVersionBundleVersion      = "VERSION_BUNDLE_VERSION"
)

func AWSAPIHostedZoneGuest() string {
	awsAPIHostedZoneGuest := os.Getenv(EnvVarAWSAPIHostedZoneGuest)
	if awsAPIHostedZoneGuest == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSAPIHostedZoneGuest))
	}
	return awsAPIHostedZoneGuest
}

func AWSIngressHostedZoneGuest() string {
	awsIngressHostedZoneGuest := os.Getenv(EnvVarAWSIngressHostedZoneGuest)
	if awsIngressHostedZoneGuest == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSIngressHostedZoneGuest))
	}
	return awsIngressHostedZoneGuest
}

func AWSRegion() string {
	awsRegion := os.Getenv(EnvVarAWSRegion)
	if awsRegion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSRegion))
	}
	return awsRegion
}

func AWSRouteTable0() string {
	awsRouteTable0 := os.Getenv(EnvVarAWSRouteTable0)
	if awsRouteTable0 == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSRouteTable0))
	}
	return awsRouteTable0
}

func AWSRouteTable1() string {
	awsRouteTable1 := os.Getenv(EnvVarAWSRouteTable1)
	if awsRouteTable1 == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSRouteTable1))
	}
	return awsRouteTable1
}

func ClusterName() string {
	clusterName := os.Getenv(EnvVarClusterName)
	if clusterName == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarClusterName))
	}
	return clusterName
}

func CommonDomain() string {
	commonDomain := os.Getenv(EnvVarCommonDomain)
	if commonDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomain))
	}
	return commonDomain
}

func SSHPublicKey() string {
	sshPublicKey := os.Getenv(EnvVarSSHPublickey)
	if sshPublicKey == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarSSHPublickey))
	}
	return sshPublicKey
}

func VersionBundleVersion() string {
	versionBundleVersion := os.Getenv(EnvVarVersionBundleVersion)
	if versionBundleVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarVersionBundleVersion))
	}
	return versionBundleVersion
}
