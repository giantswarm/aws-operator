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
			name:     "case 3: percentage - simple value",
			input:    "0.5",
			workers:  10,
			expected: "5",
		},
		{
			name:     "case 4: percentage - rounding",
			input:    "0.35",
			workers:  10,
			expected: "4",
		},
		{
			name:     "case 5: percentage - rounding",
			input:    "0.32",
			workers:  10,
			expected: "3",
		},
		{
			name:     "case 6: percentage - invalid value - too big",
			input:    "1.5",
			workers:  10,
			expected: "",
		},
		{
			name:     "case 7: percentage - invalid value - negative",
			input:    "-0.5",
			workers:  10,
			expected: "",
		},
		{
			name:     "case 8: invalid value - '50%'",
			input:    "50%",
			expected: "",
		},
		{
			name:     "case 9: invalid value - string",
			input:    "test",
			expected: "",
		},
		{
			name:     "case 10: invalid value - number and string",
			input:    "5erft",
			expected: "",
		},
		{
			name:     "case 11: invalid value - float and string",
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

func Test_PauseTimeIsValid(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "case 0: simple value",
			value: "PT15M",
			valid: true,
		},
		{
			name:  "case 1: invalid value value",
			value: "10m",
			valid: false,
		},
		{
			name:  "case 2: duration too big",
			value: "PT1H2M",
			valid: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := MachineDeploymentPauseTimeIsValid(tc.value)

			if result != tc.valid {
				t.Fatalf("%s -  expected '%t' got '%t'\n", tc.name, tc.valid, result)
			}
		})
	}
}

func Test_MachineDeploymentWorkerCountRatio(t *testing.T) {
	testCases := []struct {
		name     string
		ratio    float32
		workers  int
		expected string
	}{
		{
			name:     "case 0: simple value",
			ratio:    0.3,
			workers:  10,
			expected: "3",
		},
		{
			name:     "case 1: simple value",
			ratio:    0.9,
			workers:  10,
			expected: "9",
		},
		{
			name:     "case 2: big value",
			ratio:    0.35,
			workers:  1000,
			expected: "350",
		},
		{
			name:     "case 3: rounding",
			ratio:    0.43,
			workers:  10,
			expected: "4",
		},
		{
			name:     "case 4: rounding",
			ratio:    0.55,
			workers:  10,
			expected: "6",
		},
		{
			name:     "case 5: minimal result",
			ratio:    0.20,
			workers:  2,
			expected: "1",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := MachineDeploymentWorkerCountRatio(tc.workers, tc.ratio)

			if output != tc.expected {
				t.Fatalf("%s -  expected '%s' got '%s'\n", tc.name, tc.expected, output)
			}
		})
	}
}
