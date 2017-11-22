package patterns

import (
	"bufio"
	"io"
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Matcher interface {
	Find(input io.Reader, pattern string) (bool, error)
}

type Patterns struct {
	logger micrologger.Logger
}

func New(logger micrologger.Logger) *Patterns {
	return &Patterns{
		logger: logger,
	}
}

// Find returns true if the given pattern is found in the input pipe.
func (pa *Patterns) Find(input io.Reader, pattern string) (bool, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return false, microerror.Mask(err)
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if r.MatchString(scanner.Text()) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, microerror.Mask(err)
	}
	return false, nil
}
