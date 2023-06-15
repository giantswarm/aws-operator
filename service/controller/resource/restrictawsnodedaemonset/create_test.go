package restrictawsnodedaemonset

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/giantswarm/aws-operator/v14/pkg/project"
)

func Test_ensureAndFilterNodeSelectorRequirements(t *testing.T) {
	desiredRequirement := corev1.NodeSelectorRequirement{
		Key:      "aws-operator.giantswarm.io/version",
		Operator: "NotIn",
		Values:   []string{project.Version()},
	}
	tests := []struct {
		name            string
		requirements    []corev1.NodeSelectorRequirement
		expected        []corev1.NodeSelectorRequirement
		expectedChanged bool
	}{
		{
			name:            "case 0: empty requirements",
			requirements:    nil,
			expected:        []corev1.NodeSelectorRequirement{desiredRequirement},
			expectedChanged: true,
		},
		{
			name:            "case 1: desired requirement already in place",
			requirements:    []corev1.NodeSelectorRequirement{desiredRequirement},
			expected:        []corev1.NodeSelectorRequirement{desiredRequirement},
			expectedChanged: false,
		},
		{
			name: "case 2: some requirements in place, but desired one is not there",
			requirements: []corev1.NodeSelectorRequirement{{
				Key:      "Test",
				Operator: "In",
				Values:   []string{"test"},
			}},
			expected: []corev1.NodeSelectorRequirement{
				{
					Key:      "Test",
					Operator: "In",
					Values:   []string{"test"},
				},
				desiredRequirement,
			},
			expectedChanged: true,
		},
		{
			name: "case 3, some requirements in place, including desired one",
			requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      "Test",
					Operator: "In",
					Values:   []string{"test"},
				},
				desiredRequirement,
			},
			expected: []corev1.NodeSelectorRequirement{
				{
					Key:      "Test",
					Operator: "In",
					Values:   []string{"test"},
				},
				desiredRequirement,
			},
			expectedChanged: false,
		},
		{
			name: "case 4: wrong requirement exists",
			requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      "aws-operator.giantswarm.io/version",
					Operator: "NotIn",
					Values:   []string{"a.b.c"},
				},
			},
			expected:        []corev1.NodeSelectorRequirement{desiredRequirement},
			expectedChanged: true,
		},
		{
			name: "case 5: duplicate wrong requirement exists",
			requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      "aws-operator.giantswarm.io/version",
					Operator: "NotIn",
					Values:   []string{"a.b.c"},
				},
				{
					Key:      "aws-operator.giantswarm.io/version",
					Operator: "NotIn",
					Values:   []string{"d.e.f"},
				},
			},
			expected:        []corev1.NodeSelectorRequirement{desiredRequirement},
			expectedChanged: true,
		},
		{
			name: "case 6: wrong requirement exists among other requirements",
			requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      "Test1",
					Operator: "In",
					Values:   []string{"test"},
				},
				{
					Key:      "aws-operator.giantswarm.io/version",
					Operator: "NotIn",
					Values:   []string{"a.b.c"},
				},
				{
					Key:      "Test2",
					Operator: "In",
					Values:   []string{"test"},
				},
			},
			expected: []corev1.NodeSelectorRequirement{
				{
					Key:      "Test1",
					Operator: "In",
					Values:   []string{"test"},
				},
				{
					Key:      "Test2",
					Operator: "In",
					Values:   []string{"test"},
				},
				desiredRequirement,
			},
			expectedChanged: true,
		},
		{
			name: "case 7: wrong requirement exists alongside desired one",
			requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      "aws-operator.giantswarm.io/version",
					Operator: "NotIn",
					Values:   []string{"a.b.c"},
				},
				desiredRequirement,
			},
			expected: []corev1.NodeSelectorRequirement{
				desiredRequirement,
			},
			expectedChanged: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, changed := ensureAndFilterNodeSelectorRequirements(tt.requirements)
			if !reflect.DeepEqual(expected, tt.expected) {
				t.Errorf("ensureAndFilterNodeSelectorRequirements() expected = %v, expected %v", expected, tt.expected)
			}
			if changed != tt.expectedChanged {
				t.Errorf("ensureAndFilterNodeSelectorRequirements() changed = %v, expected %v", changed, tt.expectedChanged)
			}
		})
	}
}
