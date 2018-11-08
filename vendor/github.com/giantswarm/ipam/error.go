package ipam

import (
	"github.com/giantswarm/microerror"
)

var maskTooBigError = microerror.New("mask too big")

// IsMaskTooBig asserts maskTooBigError.
func IsMaskTooBig(err error) bool {
	return microerror.Cause(err) == maskTooBigError
}

var nilIPError = microerror.New("nil IP")

// IsNilIP asserts nilIPError.
func IsNilIP(err error) bool {
	return microerror.Cause(err) == nilIPError
}

var ipNotContainedError = microerror.New("IP not contained")

// IsIPNotContained asserts ipNotContainedError.
func IsIPNotContained(err error) bool {
	return microerror.Cause(err) == ipNotContainedError
}

var maskIncorrectSizeError = microerror.New("mask incorrect size")

// IsMaskIncorrectSize asserts maskIncorrectSizeError.
func IsMaskIncorrectSize(err error) bool {
	return microerror.Cause(err) == maskIncorrectSizeError
}

var spaceExhaustedError = microerror.New("space exhausted")

// IsSpaceExhausted asserts spaceExhaustedError.
func IsSpaceExhausted(err error) bool {
	return microerror.Cause(err) == spaceExhaustedError
}

var incorrectNumberOfBoundariesError = microerror.New("incorrect number of boundaries")

// IsIncorrectNumberOfBoundaries asserts incorrectNumberOfBoundariesError.
func IsIncorrectNumberOfBoundaries(err error) bool {
	return microerror.Cause(err) == incorrectNumberOfBoundariesError
}

var incorrectNumberOfFreeRangesError = microerror.New("incorrect number of free ranges")

// IsIncorrectNumberOfFreeRangesError asserts incorrectNumberOfFreeRangesError.
func IsIncorrectNumberOfFreeRanges(err error) bool {
	return microerror.Cause(err) == incorrectNumberOfFreeRangesError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
