package framework

import (
	"time"

	"github.com/cenkalti/backoff"
)

const (
	// REVERT
	// REVERT
	// REVERT
	// REVERT
	LongMaxWait = 80 * time.Minute
	// REVERT
	// REVERT
	// REVERT
	// REVERT
	// REVERT
	ShortMaxWait = 2 * time.Minute
)

const (
	LongMaxInterval  = 60 * time.Second
	ShortMaxInterval = 5 * time.Second
)

func NewExponentialBackoff(maxWait, maxInterval time.Duration) *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         maxInterval,
		MaxElapsedTime:      maxWait,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}
