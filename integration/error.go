package integration

import "github.com/giantswarm/microerror"

var waitTimeoutError = microerror.New("waitTimeout")

var tooManyResultsError = microerror.New("too many results")

var unexpectedStatusPhaseError = microerror.New("unexpected status phase")

var notFoundError = microerror.New("not found")
