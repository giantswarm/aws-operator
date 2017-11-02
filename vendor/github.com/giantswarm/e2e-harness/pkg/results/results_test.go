package results_test

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/results"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
)

type runnerMock struct {
	currentOutput string
}

func (r *runnerMock) Run(out io.Writer, command string) error {
	return nil
}

func (r *runnerMock) RunPortForward(out io.Writer, command string) error {
	out.Write([]byte(r.currentOutput))

	return nil
}

func TestInterpret(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll(path.Dir(results.DefaultResultsPath), os.ModePerm); err != nil {
		t.Errorf("unexepcted error creating directory %v", err)
	}

	logger := microloggertest.New()
	r := &runnerMock{}
	subject := results.New(logger, fs, r)

	testCases := []struct {
		description string
		content     string
		success     bool
	}{
		{
			description: "no failures, one test",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="1" failures="0" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
  </testsuite>
`,
			success: true,
		},
		{
			description: "no failures, multiple test",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="3" failures="0" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
  </testsuite>
`,
			success: true,
		},
		{
			description: "failure, one test",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="1" failures="1" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675">
      <failure>
      Failure message
      </failure>
    </testcase>
  </testsuite>
`,
			success: false,
		},
		{
			description: "failure, multiple tests",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="3" failures="1" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675">
      <failure>
      Failure message
      </failure>
    </testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
  </testsuite>
`,
			success: false,
		},
		{
			description: "error, one test",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="1" errors="1" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675">
      <error>
      Failure message
      </error>
    </testcase>
  </testsuite>
`,
			success: false,
		},
		{
			description: "error, multiple tests",
			content: `<?xml version="1.0" encoding="UTF-8"?>
  <testsuite tests="3" errors="1" time="16.900219675">
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675">
      <error>
      Failure message
      </error>
    </testcase>
    <testcase name="[sig-cli] Kubectl alpha client [k8s.io] Kubectl run CronJob should create a CronJob" classname="Kubernetes e2e suite" time="16.900219675"></testcase>
  </testsuite>
`,
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			r.currentOutput = tc.content

			actual := subject.Interpret()

			if tc.success && actual != nil {
				t.Errorf("unexpected result, expected success, got error %v", actual)
			}

			if !tc.success && actual == nil {
				t.Errorf("expected error didn't happen")
			}
		})
	}
}
