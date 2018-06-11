package api

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

const (
	// guestDNSNotReadyPattern is a regular expression representing DNS errors for
	// the guest API domain. Also see the following match example.
	//
	//     https://play.golang.org/p/ipBkwqlc4Td
	//
	guestDNSNotReadyPattern = "dial tcp: lookup .* on .*:53: no such host"

	// guestEOFPattern is a regular expression representing EOF errors for the
	// guest API domain. Also see the following match example.
	//
	//     https://play.golang.org/p/L6f4ItJLufv
	//
	guestEOFPattern = `Get https://api\..*/api/v1/nodes: (unexpected )?EOF`

	// guestTransientInvalidCertificatePattern regular expression defines the kind
	// of transient errors related to certificates returned while the guest API is
	// not fully up. Also see the following match example.
	//
	//     https://play.golang.org/p/iiYvBhPOg4f
	//
	guestTransientInvalidCertificatePattern = `[Get|Post] https://api\..*: x509: certificate is valid for ingress.local, not api\..*`
)

var (
	guestDNSNotReadyRegexp                 = regexp.MustCompile(guestDNSNotReadyPattern)
	guestEOFRegexp                         = regexp.MustCompile(guestEOFPattern)
	guestTransientInvalidCertificateRegexp = regexp.MustCompile(guestTransientInvalidCertificatePattern)
)

// GuestAPINotAvailableError is returned when the guest Kubernetes API is not
// available.
var GuestAPINotAvailableError = microerror.New("guest API not available")

// IsGuestAPINotAvailable asserts GuestAPINotAvailableError.
func IsGuestAPINotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	regexps := []*regexp.Regexp{
		guestDNSNotReadyRegexp,
		guestEOFRegexp,
		guestTransientInvalidCertificateRegexp,
	}
	for _, re := range regexps {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	if c == GuestAPINotAvailableError {
		return true
	}

	return false
}
