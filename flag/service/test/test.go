package test

import "github.com/giantswarm/aws-operator/flag/service/test/labelselector"

// Test aggregates flags that should be used only in order to test the project.
type Test struct {
	LabelSelector labelselector.LabelSelector
}
