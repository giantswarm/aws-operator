package tenant

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

var (
	APINotAvailablePatterns = []*regexp.Regexp{
		// A regular expression representing DNS errors for the tenant API domain.
		regexp.MustCompile(`dial tcp: lookup .* on .*:53: (no such host|server misbehaving)`),
		// Alternative DNS error appearing when running azure-operator with telepresence
		regexp.MustCompile(`[Get|Patch|Post] https://api\..* dial tcp: lookup .*: no such host`),
		// A regular expression representing EOF errors for the tenant API domain.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/api/v1/nodes.* (unexpected )?EOF`),
		// A regular expression representing EOF errors for the tenant API domain.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/api/v1/namespaces/*/.* (unexpected )?EOF`),
		// A regular expression representing TLS errors related to establishing
		// connections to tenant clusters while the tenant API is not fully up.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/api/v1/nodes.* net/http: (TLS handshake timeout|request canceled).*?`),
		// A regular expression representing timeout errors related to establishing
		// TCP connections to tenant clusters while the tenant API is not fully up.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..* dial tcp .* i/o timeout`),
		// A regular expression representing timeout errors related to awaiting headers
		// from the tenant API connection.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..* .* \(Client.Timeout exceeded while awaiting headers\)`),
		// A regular expression representing the kind of transient errors related to
		// certificates returned while the tenant API is not fully up.
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*: x509: (certificate is valid for ingress.local, not api\..*|certificate has expired or is not yet valid.*|certificate signed by unknown authority \(possibly because of "crypto/rsa: verification error" while trying to verify candidate authority certificate.*?\))`),
	}
)

// APINotAvailableError is returned when the tenant Kubernetes API is not
// available.
var APINotAvailableError = &microerror.Error{
	Kind: "APINotAvailableError",
}

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
