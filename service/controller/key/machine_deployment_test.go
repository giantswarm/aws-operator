package key

import (
	"strconv"
	"testing"
)

func Test_MachineDeploymentParseMaxBatchSize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		workers  int
		expected string
	}{
		{
			name:     "case 0: int - simple value",
			input:    "5",
			workers:  5,
			expected: "5",
		},
		{
			name:     "case 1: int - big value",
			input:    "200",
			workers:  300,
			expected: "200",
		},
		{
			name:     "case 2: int - invalid value - negative number",
			input:    "-10",
			workers:  5,
			expected: "",
		},
		{
			name:     "case 2: int - invalid value - zero",
			input:    "0",
			workers:  5,
			expected: "",
		},
		{
			name:     "case 3: int - invalid value - value bigger than worker count",
			input:    "20",
			workers:  5,
			expected: "",
		},
		{
			name:     "case 4: percentage - simple value",
			input:    "0.5",
			workers:  10,
			expected: "5",
		},
		{
			name:     "case 5: percentage - rounding",
			input:    "0.35",
			workers:  10,
			expected: "4",
		},
		{
			name:     "case 6: percentage - rounding",
			input:    "0.32",
			workers:  10,
			expected: "3",
		},
		{
			name:     "case 7: percentage - invalid value - too big",
			input:    "1.5",
			workers:  10,
			expected: "",
		},
		{
			name:     "case 8: percentage - invalid value - negative",
			input:    "-0.5",
			workers:  10,
			expected: "",
		},
		{
			name:     "case 9: invalid value - '50%'",
			input:    "50%",
			expected: "",
		},
		{
			name:     "case 10: invalid value - string",
			input:    "test",
			expected: "",
		},
		{
			name:     "case 11: invalid value - number and string",
			input:    "5erft",
			expected: "",
		},
		{
			name:     "case 12: invalid value - float and string",
			input:    "0.5erft",
			expected: "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := MachineDeploymentParseMaxBatchSize(tc.input, tc.workers)

			if output != tc.expected {
				t.Fatalf("%s -  expected '%s' got '%s'\n", tc.name, tc.expected, output)
			}
		})
	}
}