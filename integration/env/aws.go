// +build k8srequired

package env

import (
	"fmt"
	"os"
)

const (
	EnvVarAWSAPIHostedZoneGuest     = "AWS_API_HOSTED_ZONE_GUEST"
	EnvVarAWSIngressHostedZoneGuest = "AWS_INGRESS_HOSTED_ZONE_GUEST"
	EnvVarAWSRegion                 = "AWS_REGION"
	EnvVarGuestAWSARN               = "GUEST_AWS_ARN"
	EnvVarGuestAWSAccessKeyID       = "GUEST_AWS_ACCESS_KEY_ID"
	EnvVarGuestAWSAccessKeySecret   = "GUEST_AWS_SECRET_ACCESS_KEY"
	EnvVarHostAWSAccessKeyID        = "HOST_AWS_ACCESS_KEY_ID"
	EnvVarHostAWSAccessKeySecret    = "HOST_AWS_SECRET_ACCESS_KEY"
	EnvVarIDRSAPub                  = "IDRSA_PUB"
)

var (
	awsAPIHostedZoneGuest     string
	awsIngressHostedZoneGuest string
	awsRegion                 string
	guestAWSARN               string
	guestAWSAccessKeyID       string
	guestAWSAccessKeySecret   string
	hostAWSAccessKeyID        string
	hostAWSAccessKeySecret    string
	idRSAPub                  string
)

func init() {
	awsAPIHostedZoneGuest = os.Getenv(EnvVarAWSAPIHostedZoneGuest)
	if awsAPIHostedZoneGuest == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSAPIHostedZoneGuest))
	}

	awsIngressHostedZoneGuest = os.Getenv(EnvVarAWSIngressHostedZoneGuest)
	if awsIngressHostedZoneGuest == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSIngressHostedZoneGuest))
	}

	awsRegion = os.Getenv(EnvVarAWSRegion)
	if awsRegion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarAWSRegion))
	}

	guestAWSARN = os.Getenv(EnvVarGuestAWSARN)
	if guestAWSARN == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarGuestAWSARN))
	}

	guestAWSAccessKeyID = os.Getenv(EnvVarGuestAWSAccessKeyID)
	if guestAWSAccessKeyID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarGuestAWSAccessKeyID))
	}

	guestAWSAccessKeySecret = os.Getenv(EnvVarGuestAWSAccessKeySecret)
	if guestAWSAccessKeySecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarGuestAWSAccessKeySecret))
	}

	hostAWSAccessKeyID = os.Getenv(EnvVarHostAWSAccessKeyID)
	if hostAWSAccessKeyID == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarHostAWSAccessKeyID))
	}

	hostAWSAccessKeySecret = os.Getenv(EnvVarHostAWSAccessKeySecret)
	if hostAWSAccessKeySecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarHostAWSAccessKeySecret))
	}

	idRSAPub = os.Getenv(EnvVarIDRSAPub)
	if idRSAPub == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarIDRSAPub))
	}
}

func AWSAPIHostedZoneGuest() string {
	return awsAPIHostedZoneGuest
}

func AWSIngressHostedZoneGuest() string {
	return awsIngressHostedZoneGuest
}

func AWSRegion() string {
	return awsRegion
}

func AWSRouteTable0() string {
	return ClusterID() + "_0"
}

func AWSRouteTable1() string {
	return ClusterID() + "_1"
}

func GuestAWSARN() string {
	return guestAWSARN
}

func GuestAWSAccessKeyID() string {
	return guestAWSAccessKeyID
}

func GuestAWSAccessKeySecret() string {
	return guestAWSAccessKeySecret
}

func GuestAWSAccessKeyToken() string {
	return ""
}

func HostAWSAccessKeyID() string {
	return hostAWSAccessKeyID
}

func HostAWSAccessKeySecret() string {
	return hostAWSAccessKeySecret
}

func HostAWSAccessKeyToken() string {
	return ""
}

func IDRSAPub() string {
	return idRSAPub
}
