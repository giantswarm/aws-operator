package create

import (
	"fmt"
	"testing"

	"github.com/juju/errgo"
	"github.com/stretchr/testify/assert"
)

func TestHostedZoneName(t *testing.T) {
	tests := []struct {
		desc   string
		domain string
		res    string
		err    error
	}{
		{
			desc:   "well-formed domain",
			domain: "api.foobar.example.customer.com",
			res:    "example.customer.com",
		},
		{
			desc:   "another well-formed domain",
			domain: "this.is.a.well.formed.domain",
			res:    "a.well.formed.domain",
		},
		{
			desc:   "empty domain",
			domain: "",
			res:    "",
			err:    malformedCloudConfigKeyError,
		},
		{
			desc:   "malformed domain",
			domain: "not a domain",
			res:    "",
			err:    malformedCloudConfigKeyError,
		},
	}

	for _, tc := range tests {
		res, err := hostedZoneName(tc.domain)

		if err != nil {
			underlying := errgo.Cause(err)
			assert.Equal(t, tc.err, underlying, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
		}

		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
	}
}
