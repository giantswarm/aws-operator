package guest

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

var (
	APINotAvailablePatterns = []*regexp.Regexp{
		// A regular expression representing DNS errors for the guest API domain.
		regexp.MustCompile(`dial tcp: lookup .* on .*:53: (no such host|server misbehaving)`),
		// A regular expression representing EOF errors for the guest API domain.
		regexp.MustCompile(`Get https://api\..*/api/v1/nodes.* (unexpected )?EOF`),
		// A regular expression representing EOF errors for the guest API domain.
		regexp.MustCompile(`[Get|Post] https://api\..*/api/v1/namespaces/*/.* (unexpected )?EOF`),
		// A regular expression representing TLS errors related to establishing
		// connections to guest clusters while the guest API is not fully up.
		regexp.MustCompile(`Get https://api\..*/api/v1/nodes.* net/http: (TLS handshake timeout|request canceled while waiting for connection).*?`),
		// A regular expression representing the kind of transient errors related to
		// certificates returned while the guest API is not fully up.
		regexp.MustCompile(`[Get|Post] https://api\..*: x509: (certificate is valid for ingress.local, not api\..*|certificate signed by unknown authority \(possibly because of "crypto/rsa: verification error" while trying to verify candidate authority certificate.*?\))`),
	}
)

// APINotAvailableError is returned when the guest Kubernetes API is not
// available.
var APINotAvailableError = microerror.New("API not available")

// IsAPINotAvailable asserts APINotAvailableError.
func IsAPINotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	for _, re := range APINotAvailablePatterns {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	if c == APINotAvailableError {
		return true
	}

	return false
}
