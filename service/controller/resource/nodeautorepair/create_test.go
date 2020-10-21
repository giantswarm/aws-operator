package nodeautorepair

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
		expectedError      bool
	}{
		{
			name: "test 0 - basic test",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-west-1c/i-06a1d2fe9b3e8c916",
				},
			},
			expectedInstanceID: "i-06a1d2fe9b3e8c916",
			expectedError:      false,
		},
		{
			name: "test 1 - basic test",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///cn-north-1c/i-a1dasd1wddas3e8c916",
				},
			},
			expectedInstanceID: "i-a1dasd1wddas3e8c916",
			expectedError:      false,
		},
		{
			name: "test 2 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-west-1c/",
				},
			},
			expectedInstanceID: "",
			expectedError:      true,
		},
		{
			name: "test 3 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "",
				},
			},
			expectedInstanceID: "",
			expectedError:      true,
		},
		{
			name: "test 3 - bad provider ID",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws://cn-north-1c/i-a1dasd1wddas3e8c916",
				},
			},
			expectedInstanceID: "",
			expectedError:      true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			instanceID, err := getInstanceId(tc.node)
			if err != nil && !tc.expectedError {
				t.Fatalf("Encountered an error when parsing valid providerID '%s'.\n", tc.node.Spec.ProviderID)
			}

			if err == nil && tc.expectedError {
				t.Fatalf("Expected an error during parsiong of instanceID '%s' but there was none.\n", tc.node.Spec.ProviderID)
			}

			if instanceID != tc.expectedInstanceID {
				t.Fatalf("Expected '%s' instance id but got '%s'.\n", tc.expectedInstanceID, instanceID)
			}
		})
	}
}
