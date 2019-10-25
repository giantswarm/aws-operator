package helmclient

import "strings"

// isNoSuchHostError asserts no such host error.
func isNoSuchHostError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "no such host")
}
