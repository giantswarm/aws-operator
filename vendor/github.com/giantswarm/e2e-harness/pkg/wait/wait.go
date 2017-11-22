package wait

import (
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/e2e-harness/pkg/patterns"
	"github.com/giantswarm/e2e-harness/pkg/runner"
)

const (
	defaultDeadline = 120000
	defaultStep     = 500
)

type Wait struct {
	logger  micrologger.Logger
	matcher patterns.Matcher
	runner  runner.Runner
}

type MatchDef struct {
	Run   string
	Match string
	// total time to wait
	Deadline time.Duration
	// delay between checks
	Step time.Duration
}

func New(logger micrologger.Logger, runner runner.Runner, matcher patterns.Matcher) *Wait {
	return &Wait{
		logger:  logger,
		runner:  runner,
		matcher: matcher,
	}
}

func (w *Wait) For(md *MatchDef) error {
	if md.Deadline == 0 {
		md.Deadline = defaultDeadline
	}
	if md.Step == 0 {
		md.Step = defaultStep
	}

	timeout := time.After(md.Deadline * time.Millisecond)
	tick := time.Tick(md.Step * time.Millisecond)

	w.logger.Log("debug", fmt.Sprintf("waiting for pattern %q in %q output", md.Match, md.Run))
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout looking for pattern %q with command %q", md.Match, md.Run)
		case <-tick:
			// pipe the output of the docker command to the input of FindMatch,
			// this way we can handle potentially long outputs without having
			// to store them in a variable
			re, wr := io.Pipe()
			// writing without a reader will deadlock so write in a goroutine
			go func() {
				defer wr.Close()
				w.runner.RunPortForward(wr, md.Run)
			}()
			ok, err := w.matcher.Find(re, md.Match)
			if err != nil {
				return microerror.Mask(err)
			} else if ok {
				return nil
			}
		}
	}
}
