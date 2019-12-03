package v1alpha1

import (
	"testing"
	"time"
)

func Test_HasDrainedCondition(t *testing.T) {
	testCases := []struct {
		name           string
		status         DrainerConfigStatus
		expectedResult bool
	}{
		{
			name:           "case 0: DrainerConfigStatus with empty Conditions doesn't have Drained condition",
			status:         DrainerConfigStatus{},
			expectedResult: false,
		},
		{
			name: "case 1: DrainerConfigStatus with Drained status condition in conditions",
			status: DrainerConfigStatus{
				Conditions: []DrainerConfigStatusCondition{
					DrainerConfigStatusCondition{
						LastTransitionTime: DeepCopyTime{time.Now()},
						Status:             DrainerConfigStatusStatusTrue,
						Type:               DrainerConfigStatusTypeDrained,
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 2: DrainerConfigStatus with Timeout status condition in conditions doesn't have Drained condition",
			status: DrainerConfigStatus{
				Conditions: []DrainerConfigStatusCondition{
					DrainerConfigStatusCondition{
						LastTransitionTime: DeepCopyTime{time.Now()},
						Status:             DrainerConfigStatusStatusTrue,
						Type:               DrainerConfigStatusTypeTimeout,
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := tc.status.HasDrainedCondition()

			if h != tc.expectedResult {
				t.Fatalf("HasDrainedCondition() == %v, expected %v", h, tc.expectedResult)
			}
		})
	}
}

func Test_HasTimeoutCondition(t *testing.T) {
	testCases := []struct {
		name           string
		status         DrainerConfigStatus
		expectedResult bool
	}{
		{
			name:           "case 0: DrainerConfigStatus with empty Conditions doesn't have Timeout condition",
			status:         DrainerConfigStatus{},
			expectedResult: false,
		},
		{
			name: "case 1: DrainerConfigStatus with Timeout status condition in conditions",
			status: DrainerConfigStatus{
				Conditions: []DrainerConfigStatusCondition{
					DrainerConfigStatusCondition{
						LastTransitionTime: DeepCopyTime{time.Now()},
						Status:             DrainerConfigStatusStatusTrue,
						Type:               DrainerConfigStatusTypeTimeout,
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 2: DrainerConfigStatus with Drained status condition in conditions doesn't have Timeout condition",
			status: DrainerConfigStatus{
				Conditions: []DrainerConfigStatusCondition{
					DrainerConfigStatusCondition{
						LastTransitionTime: DeepCopyTime{time.Now()},
						Status:             DrainerConfigStatusStatusTrue,
						Type:               DrainerConfigStatusTypeDrained,
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := tc.status.HasTimeoutCondition()

			if h != tc.expectedResult {
				t.Fatalf("HasTimeoutCondition() == %v, expected %v", h, tc.expectedResult)
			}
		})
	}
}

func Test_NewDrainedCondition(t *testing.T) {
	status := DrainerConfigStatus{}
	status.Conditions = append(status.Conditions, status.NewDrainedCondition())

	if !status.HasDrainedCondition() {
		t.Fatalf("DrainerConfigStatus doesn't have Drained condition after NewDrainedCondition() call")
	}
}

func Test_NewTimeoutCondition(t *testing.T) {
	status := DrainerConfigStatus{}
	status.Conditions = append(status.Conditions, status.NewTimeoutCondition())

	if !status.HasTimeoutCondition() {
		t.Fatalf("DrainerConfigStatus doesn't have Timeout condition after NewTimeoutCondition() call")
	}
}
