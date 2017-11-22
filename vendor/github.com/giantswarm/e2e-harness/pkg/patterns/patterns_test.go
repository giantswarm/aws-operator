package patterns_test

import (
	"strings"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/e2e-harness/pkg/patterns"
)

func TestFindMatch(t *testing.T) {
	testCases := []struct {
		description string
		input       string
		pattern     string
		error       bool
		expected    bool
	}{
		{
			description: "basic match",
			input:       "aabbaa",
			pattern:     "bb",
			error:       false,
			expected:    true,
		},
		{
			description: "multiline match",
			input: `aa
bb
aa`,
			pattern:  "bb",
			error:    false,
			expected: true,
		},
		{
			description: "kubectl-like match",
			input: `NAMESPACE     NAME                             READY     STATUS    RESTARTS   AGE
kube-system   kube-addon-manager-minikube      1/1       Running   1          16h
kube-system   kube-dns-1326421443-whr1s        3/3       Running   3          16h
kube-system   kubernetes-dashboard-29pll       1/1       Running   1          16h
kube-system   tiller-deploy-1046433508-ch23h   1/1       Running   1          16h
`,
			pattern:  `tiller-deploy.*1/1\s*Running`,
			error:    false,
			expected: true,
		},
		{
			description: "basic unmatch",
			input:       "aaccaa",
			pattern:     "bb",
			error:       false,
			expected:    false,
		},
		{
			description: "multiline unmatch",
			input: `aa
cc
aa`,
			pattern:  "bb",
			error:    false,
			expected: false,
		},
		{
			description: "kubectl-like unmatch",
			input: `NAMESPACE     NAME                             READY     STATUS    RESTARTS   AGE
kube-system   kube-addon-manager-minikube      1/1       Running   1          16h
kube-system   kube-dns-1326421443-whr1s        3/3       Running   3          16h
kube-system   kubernetes-dashboard-29pll       1/1       Running   1          16h
kube-system   tiller-deploy-1046433508-ch23h   1/1       Running   1          16h
`,
			pattern:  `mycustom-deploy.*1/1\s*Running`,
			error:    false,
			expected: false,
		},
		{
			description: "pattern error",
			input:       "aabbaa",
			pattern:     "a([a-z",
			error:       true,
			expected:    false,
		},
	}
	logger := microloggertest.New()
	subject := patterns.New(logger)

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			inputPipe := strings.NewReader(tc.input)
			actual, err := subject.Find(inputPipe, tc.pattern)

			if tc.error && err == nil {
				t.Errorf("expected error didn't happen")
			}

			if !tc.error && err != nil {
				t.Errorf("unexpected error %q", err)
			}

			if actual != tc.expected {
				t.Errorf("match of %q with %q failed, expected %t and found %t",
					tc.input, tc.pattern, tc.expected, actual)
			}
		})
	}
}
