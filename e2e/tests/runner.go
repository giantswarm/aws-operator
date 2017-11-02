package tests

import (
	"github.com/giantswarm/e2e-harness/pkg/results"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
)

// Test is a generic type for all the test functions, it returns a
// description of the tests and the eventually returned error
type Test func() (string, error)

type TestSet struct {
	clientset *kubernetes.Clientset
	logger    micrologger.Logger
}

var (
	// tests holds the array of functions to be executed
	tests = []Test{}
	// ts is the test suite that will keep the results
	ts = &results.TestSuite{}
)

// Run executes all the tests and saves the results
func Run() error {
	for _, test := range tests {
		ts.Tests++
		desc, err := test()
		tc := results.TestCase{
			Name: desc,
		}
		if err != nil {
			ts.Failures++
			tc.Failure = &results.TestFailure{
				Value: err.Error(),
			}
		}
		ts.TestCases = append(ts.TestCases, tc)
	}
	fs := afero.NewOsFs()
	return results.Write(fs, ts)
}

// Add appends the given test to the existing bundle
func Add(t Test) {
	tests = append(tests, t)
}
