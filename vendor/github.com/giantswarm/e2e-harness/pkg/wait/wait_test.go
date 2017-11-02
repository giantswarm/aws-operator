package wait_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/giantswarm/e2e-harness/pkg/wait"
	"github.com/giantswarm/micrologger/microloggertest"
)

type runnerMock struct{}

func (r *runnerMock) Run(out io.Writer, command string) error {
	return nil
}

func (r *runnerMock) RunPortForward(out io.Writer, command string) error {
	return nil
}

type matcherMock struct {
	untilMatch int
	current    int
	doError    bool
}

func (p *matcherMock) Find(input io.Reader, pattern string) (bool, error) {
	if p.doError {
		return false, fmt.Errorf("error from pattern matching!")
	}
	if p.current < p.untilMatch {
		p.current++
		return false, nil
	}
	return true, nil
}

func TestWait(t *testing.T) {
	testCases := []struct {
		description       string
		deadline          time.Duration
		step              time.Duration
		untilMatch        int
		success           bool
		patternMatchError bool
	}{
		{
			description: "Basic match",
			deadline:    10,
			step:        1,
			untilMatch:  5,
			success:     true,
		},
		{
			description: "Basic unmatch, timeout",
			deadline:    10,
			step:        1,
			untilMatch:  20,
			success:     false,
		},
		{
			description:       "Basic unmatch, pattern match error",
			deadline:          10,
			step:              1,
			untilMatch:        5,
			success:           false,
			patternMatchError: true,
		},
	}

	logger := microloggertest.New()
	runner := &runnerMock{}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			matcher := &matcherMock{
				untilMatch: tc.untilMatch,
				doError:    tc.patternMatchError,
			}
			subject := wait.New(logger, runner, matcher)
			md := &wait.MatchDef{
				Deadline: tc.deadline,
				Step:     tc.step,
			}
			actual := subject.For(md)
			if tc.success && actual != nil {
				t.Errorf("expected success, error returned %v", actual)
			}
			if !tc.success && actual == nil {
				t.Errorf("expected error didn't happen")
			}
		})
	}
}
