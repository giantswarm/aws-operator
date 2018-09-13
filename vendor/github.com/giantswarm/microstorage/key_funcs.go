package microstorage

import (
	"strings"

	"github.com/giantswarm/microerror"
)

// SanitizeListKey sanitizes key ad makes sure it is valid for the Storage.List
// operation. List is special because unlike other operations it can take "/"
// as a key. Otherwise this is behaves like SanitizeKey function.
func SanitizeListKey(key string) (string, error) {
	if key == "/" {
		return key, nil
	}
	return SanitizeKey(key)
}

// SanitizeKey ensures the key has leading slash and does not have trailing
// slash. It fails with InvalidKeyError when key is invalid.
//
// A valid key does not contain double slashes, is not empty, and does not
// contain only slashes.
//
// This function is meant to be used by Storage implementations to simplify key
// validation logic and potentially implementation logic, because it reduces
// number of invariants to check greatly.
//
// NOTE: There is a special case of this function: SanitizeListKey.
// isValidKey checks if this storage key is valid, i.e. does not contain double
// slashes, is not empty, and does not contain only slashes.
func SanitizeKey(key string) (string, error) {
	// Make sure the key is valid.
	if key == "" || key == "/" || strings.Contains(key, "//") {
		return "", microerror.Maskf(InvalidKeyError, "key=%s", key)
	}

	// Ensure leading slash and no trailing slash.
	if key[0] != '/' {
		key = "/" + key
	}
	if key[len(key)-1] == '/' {
		key = key[:len(key)-1]
	}

	return key, nil
}
