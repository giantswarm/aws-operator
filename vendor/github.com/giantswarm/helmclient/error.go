package helmclient

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	cannotReuseReleaseErrorPrefix = "cannot re-use"
	// guestDNSNotReadyPattern is a regular expression representing DNS errors for
	// the guest API domain.
	// match example https://play.golang.org/p/ipBkwqlc4Td
	guestDNSNotReadyPattern = "dial tcp: lookup .* on .*:53: no such host"

	// guestTransientInvalidCertificatePattern regular expression defines the kind
	// of transient errors related to certificates returned while the guest API is
	// not fully up.
	// match example https://play.golang.org/p/iiYvBhPOg4f
	guestTransientInvalidCertificatePattern = `[Get|Post] https://api\..*: x509: certificate is valid for ingress.local, not api\..*`
)

var (
	guestDNSNotReadyRegexp                 *regexp.Regexp
	guestTransientInvalidCertificateRegexp *regexp.Regexp
)

func init() {
	guestDNSNotReadyRegexp = regexp.MustCompile(guestDNSNotReadyPattern)
	guestTransientInvalidCertificateRegexp = regexp.MustCompile(guestTransientInvalidCertificatePattern)
}

var cannotReuseReleaseError = microerror.New("cannot reuse release")

// IsCannotReuseRelease asserts cannotReuseReleaseError.
func IsCannotReuseRelease(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.Contains(c.Error(), cannotReuseReleaseErrorPrefix) {
		return true
	}
	if c == cannotReuseReleaseError {
		return true
	}

	return false
}

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var guestAPINotAvailableError = microerror.New("Guest API not available")
var guestNamespaceCreationErrorSuffix = "namespaces/kube-system/serviceaccounts: EOF"

// IsGuestAPINotAvailable asserts guestAPINotAvailableError.
func IsGuestAPINotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.HasSuffix(c.Error(), guestNamespaceCreationErrorSuffix) {
		return true
	}

	regexps := []*regexp.Regexp{
		guestDNSNotReadyRegexp,
		guestTransientInvalidCertificateRegexp,
	}
	for _, re := range regexps {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	if c == guestAPINotAvailableError {
		return true
	}

	return false
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

const (
	invalidGZipHeaderErrorPrefix = "gzip: invalid header"
)

var invalidGZipHeaderError = microerror.New("invalid gzip header")

// IsInvalidGZipHeader asserts invalidGZipHeaderError.
func IsInvalidGZipHeader(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.HasPrefix(c.Error(), invalidGZipHeaderErrorPrefix) {
		return true
	}
	if c == invalidGZipHeaderError {
		return true
	}

	return false
}

var podNotFoundError = microerror.New("pod not found")

// IsPodNotFound asserts podNotFoundError.
func IsPodNotFound(err error) bool {
	return microerror.Cause(err) == podNotFoundError
}

const (
	releaseNotFoundErrorPrefix = "No such release:"
	releaseNotFoundErrorSuffix = "not found"
)

var releaseNotFoundError = microerror.New("release not found")

// IsReleaseNotFound asserts releaseNotFoundError.
func IsReleaseNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.HasPrefix(c.Error(), releaseNotFoundErrorPrefix) {
		return true
	}
	if strings.HasSuffix(c.Error(), releaseNotFoundErrorSuffix) {
		return true
	}
	if c == releaseNotFoundError {
		return true
	}

	return false
}

var tillerInstallationFailedError = microerror.New("Tiller installation failed")

// IsTillerInstallationFailed asserts tillerInstallationFailedError.
func IsTillerInstallationFailed(err error) bool {
	return microerror.Cause(err) == tillerInstallationFailedError
}

var tooManyResultsError = microerror.New("too many results")

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
