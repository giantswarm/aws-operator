package terminateunhealthynode

import (
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func Test_getInstanceId(t *testing.T) {
	testCases := []struct {
		name               string
		node               corev1.Node
		expectedInstanceID string
		errorMatcher       func(error) bool
	}{
		{
			name: "test 0 - basic test",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-west-1c/i-06a1d2fe9b3e8c916",
				},
			},
			expectedInstanceID: "i-06a1d2fe9b3e8c916",
			errorMatcher:       nil,
		},
		{
			name: "test 1 - basic test",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///cn-north-1c/i-a1dasd1wddas3e8c916",
				},
			},
			expectedInstanceID: "i-a1dasd1wddas3e8c916",
			errorMatcher:       nil,
		},
		{
			name: "test 2 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-west-1c/",
				},
			},
			expectedInstanceID: "",
			errorMatcher:       IsInvalidProviderID,
		},
		{
			name: "test 3 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "",
				},
			},
			expectedInstanceID: "",
			errorMatcher:       IsInvalidProviderID,
		},
		{
			name: "test 3 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws://cn-north-1c/i-a1dasd1wddas3e8c916",
				},
			},
			expectedInstanceID: "",
			errorMatcher:       IsInvalidProviderID,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			instanceID, err := getInstanceId(tc.node)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if instanceID != tc.expectedInstanceID {
				t.Fatalf("Expected '%s' instance id but got '%s'.\n", tc.expectedInstanceID, instanceID)
			}
		})
	}
}
