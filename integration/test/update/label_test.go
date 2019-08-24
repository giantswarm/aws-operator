package update

import (
	"testing"
)

func Test_Update_Provider_ensureLabel(t *testing.T) {
	testCases := []struct {
		Name           string
		Labels         string
		Key            string
		Value          string
		ExpectedLabels string
	}{
		{
			Name:           "case 0",
			Labels:         "",
			Key:            "",
			Value:          "",
			ExpectedLabels: "",
		},
		{
			Name:           "case 1",
			Labels:         "",
			Key:            "version",
			Value:          "1.0.0",
			ExpectedLabels: "version=1.0.0",
		},
		{
			Name:           "case 2",
			Labels:         "ip=127.0.0.1",
			Key:            "version",
			Value:          "1.0.0",
			ExpectedLabels: "ip=127.0.0.1,version=1.0.0",
		},
		{
			Name:           "case 3",
			Labels:         "host=whatever,ip=127.0.0.1",
			Key:            "version",
			Value:          "1.0.0",
			ExpectedLabels: "host=whatever,ip=127.0.0.1,version=1.0.0",
		},
		{
			Name:           "case 3",
			Labels:         "host=whatever,ip=127.0.0.1,version=1.0.0",
			Key:            "version",
			Value:          "2.0.0",
			ExpectedLabels: "host=whatever,ip=127.0.0.1,version=2.0.0",
		},
		{
			Name:           "case 4",
			Labels:         "host=whatever,ip=127.0.0.1,version=1.0.0",
			Key:            "kvm-operator.giantswarm.io/version",
			Value:          "2.0.0",
			ExpectedLabels: "host=whatever,ip=127.0.0.1,version=1.0.0,kvm-operator.giantswarm.io/version=2.0.0",
		},
		{
			Name:           "case 5",
			Labels:         "host=whatever,ip=127.0.0.1,version=1.0.0,kvm-operator.giantswarm.io/version=2.0.0",
			Key:            "kvm-operator.giantswarm.io/version",
			Value:          "3.5.16",
			ExpectedLabels: "host=whatever,ip=127.0.0.1,version=1.0.0,kvm-operator.giantswarm.io/version=3.5.16",
		},
	}

	for _, tc := range testCases {
		labels := ensureLabel(tc.Labels, tc.Key, tc.Value)
		if labels != tc.ExpectedLabels {
			t.Fatalf("expected '%s' got '%s'", tc.ExpectedLabels, labels)
		}
	}
}
