package env

import (
	"fmt"
	"os"
)

const (
	EnvVarCommonDomain = "COMMON_DOMAIN"
	EnvVarVaultToken   = "VAULT_TOKEN"
)

var (
	commonDomain string
	vaultToken   string
)

func init() {
	commonDomain = os.Getenv(EnvVarCommonDomain)
	if commonDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomain))
	}

	vaultToken = os.Getenv(EnvVarVaultToken)
	if vaultToken == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarVaultToken))
	}
}

func CommonDomain() string {
	return commonDomain
}

func VaultToken() string {
	return vaultToken
}
