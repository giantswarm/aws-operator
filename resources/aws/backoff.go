package aws

import (
	"time"

	"github.com/cenkalti/backoff"
)

const maxElapsedTime = 2 * time.Minute

// NewCustomExponentialBackoff returns pointer to a backoff.ExponentialBackOff,
// initialized with custom values. At the moment, we only override the
// MaxElapsedTime.
func NewCustomExponentialBackoff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      maxElapsedTime,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}
