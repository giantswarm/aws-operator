package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_DeepCopyDate_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name         string
		inputDate    string
		expectedJSON string
		errorMatcher func(err error) bool
	}{
		{
			name:         "case 0: valid date",
			inputDate:    "2019-04-05T00:00:00Z",
			expectedJSON: `{"testDate":"2019-04-05"}`,
			errorMatcher: nil,
		},
		{
			name:         "case 1: valid date, non-zeroed hour",
			inputDate:    "2019-04-05T12:05:17Z",
			expectedJSON: `{"testDate":"2019-04-05"}`,
			errorMatcher: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(time.RFC3339, tc.inputDate)
			if !reflect.DeepEqual(nil, err) {
				t.Fatalf("\n\n%s\n", cmp.Diff(nil, err))
			}

			vPointer := struct {
				TestDate *DeepCopyDate `json:"testDate"`
			}{
				TestDate: &DeepCopyDate{
					Time: date,
				},
			}

			vValue := struct {
				TestDate DeepCopyDate `json:"testDate"`
			}{
				TestDate: DeepCopyDate{
					Time: date,
				},
			}

			bytesPointer, err := json.Marshal(vPointer)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %v, want matching", err)
			}

			bytesValue, err := json.Marshal(vValue)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %v, want matching", err)
			}

			if tc.errorMatcher != nil {
				return
			}

			if !reflect.DeepEqual(string(bytesPointer), tc.expectedJSON) {
				t.Fatalf("\n\n%s\n", cmp.Diff(string(bytesPointer), tc.expectedJSON))
			}

			if !reflect.DeepEqual(string(bytesValue), tc.expectedJSON) {
				t.Fatalf("\n\n%s\n", cmp.Diff(string(bytesValue), tc.expectedJSON))
			}
		})
	}
}

func Test_DeepCopyDate_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name          string
		inputJSON     string
		expectedNil   bool
		expectedDay   int
		expectedMonth time.Month
		expectedYear  int
		errorMatcher  func(err error) bool
	}{
		{
			name:          "case 0: valid date",
			inputJSON:     `{"testDate":"2019-02-08"}`,
			expectedNil:   false,
			expectedDay:   8,
			expectedMonth: 2,
			expectedYear:  2019,
			errorMatcher:  nil,
		},
		{
			name:         "case 1: malformed date",
			inputJSON:    `{"testDate":"2019-02-08T12:04:00"}`,
			errorMatcher: func(err error) bool { return err != nil },
		},
		{
			name:          "case 2: null",
			inputJSON:     `{"testDate":null}`,
			expectedNil:   true,
			expectedDay:   1,
			expectedMonth: 1,
			expectedYear:  1,
			errorMatcher:  nil,
		},
		{
			name:         "case 3: wrong type",
			inputJSON:    `{"testDate":5}`,
			errorMatcher: func(err error) bool { return err != nil },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			type JSONObjectPointer struct {
				TestDate *DeepCopyDate `json:"testDate"`
			}

			type JSONObjectValue struct {
				TestDate DeepCopyDate `json:"testDate"`
			}

			var jsonObjectPointer = JSONObjectPointer{TestDate: &DeepCopyDate{}}
			err := json.Unmarshal([]byte(tc.inputJSON), &jsonObjectPointer)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %v, want matching", err)
			}

			var jsonObjectValue JSONObjectValue
			err = json.Unmarshal([]byte(tc.inputJSON), &jsonObjectValue)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %v, want matching", err)
			}

			if tc.errorMatcher != nil {
				return
			}

			if tc.expectedNil {
				// We need a nil of the same type. Otherwise won't work with cmp.
				var nilDeepCopyDate *DeepCopyDate
				if !reflect.DeepEqual(jsonObjectPointer.TestDate, nilDeepCopyDate) {
					t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectPointer.TestDate, nilDeepCopyDate))
				}
			} else {
				if !reflect.DeepEqual(jsonObjectPointer.TestDate.Day(), tc.expectedDay) {
					t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectPointer.TestDate.Day(), tc.expectedDay))
				}
				if !reflect.DeepEqual(jsonObjectPointer.TestDate.Month(), tc.expectedMonth) {
					t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectPointer.TestDate.Month(), tc.expectedMonth))
				}
				if !reflect.DeepEqual(jsonObjectPointer.TestDate.Year(), tc.expectedYear) {
					t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectPointer.TestDate.Year(), tc.expectedYear))
				}
			}

			if !reflect.DeepEqual(jsonObjectValue.TestDate.Day(), tc.expectedDay) {
				t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectValue.TestDate.Day(), tc.expectedDay))
			}
			if !reflect.DeepEqual(jsonObjectValue.TestDate.Month(), tc.expectedMonth) {
				t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectValue.TestDate.Month(), tc.expectedMonth))
			}
			if !reflect.DeepEqual(jsonObjectValue.TestDate.Year(), tc.expectedYear) {
				t.Errorf("\n\n%s\n", cmp.Diff(jsonObjectValue.TestDate.Year(), tc.expectedYear))
			}
		})
	}
}
