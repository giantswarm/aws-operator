package changedetection

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ChangeDetection_TCNP_securityGroupsEqual(t *testing.T) {
	testCases := []struct {
		name    string
		current []string
		desired []string
		result  bool
	}{
		{
			name:    "case 0",
			current: nil,
			desired: nil,
			result:  true,
		},
		{
			name:    "case 1",
			current: []string{},
			desired: []string{},
			result:  true,
		},
		{
			name: "case 2",
			current: []string{
				"a",
			},
			desired: []string{},
			result:  false,
		},
		{
			name: "case 3",
			current: []string{
				"a",
			},
			desired: []string{
				"a",
			},
			result: true,
		},
		{
			name: "case 4",
			current: []string{
				"a",
				"b",
				"c",
			},
			desired: []string{
				"a",
				"b",
				"c",
			},
			result: true,
		},
		{
			name: "case 5",
			current: []string{
				"b",
				"a",
				"c",
			},
			desired: []string{
				"c",
				"a",
				"b",
			},
			result: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := securityGroupsEqual(tc.current, tc.desired)

			if result != tc.result {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.result, result))
			}
		})
	}
}
