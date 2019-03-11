package vault

import (
	"reflect"
	"testing"
)

func Test_stringSlice_Add(t *testing.T) {
	testCases := []struct {
		name          string
		slice         []string
		add           []string
		expectedSlice []string
	}{
		{
			name:          "case 0",
			slice:         []string{"a", "b", "c"},
			add:           []string{"d", "e"},
			expectedSlice: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:          "case 1",
			slice:         []string{"a", "b", "c"},
			add:           []string{"d", "a"},
			expectedSlice: []string{"a", "b", "c", "d"},
		},
		{
			name:          "case 2",
			slice:         []string{"a", "b", "c"},
			add:           []string{"b", "c"},
			expectedSlice: []string{"a", "b", "c"},
		},
		{
			name:          "case 3",
			slice:         []string{},
			add:           []string{"a", "b"},
			expectedSlice: []string{"a", "b"},
		},
		{
			name:          "case 4",
			slice:         []string{"a", "b"},
			add:           []string{},
			expectedSlice: []string{"a", "b"},
		},
		{
			name:          "case 5",
			slice:         []string{},
			add:           []string{},
			expectedSlice: []string{},
		},
		{
			name:          "case 6",
			slice:         []string{"a", "b"},
			add:           []string{"a", "b", "b", "a"},
			expectedSlice: []string{"a", "b"},
		},
		{
			name:          "case 7",
			slice:         []string{"a"},
			add:           []string{"b", "b", "b", "c"},
			expectedSlice: []string{"a", "b", "c"},
		},
		{
			name:          "case 8",
			slice:         []string{"a", "a", "a", "b"},
			add:           []string{"a", "c"},
			expectedSlice: []string{"a", "a", "a", "b", "c"},
		},
		{
			name:          "case 9",
			slice:         []string{"a", "a", "a", "b"},
			add:           []string{},
			expectedSlice: []string{"a", "a", "a", "b"},
		},
		{
			name:          "case 10",
			slice:         []string{"a"},
			add:           []string{"a", "a", "a", "b", "b"},
			expectedSlice: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		slice := stringSlice(tc.slice).Add(tc.add...)

		if !reflect.DeepEqual(slice, tc.expectedSlice) {
			t.Fatalf("slice == %v, want %v", slice, tc.expectedSlice)
		}
	}
}

func Test_stringSlice_Delete(t *testing.T) {
	testCases := []struct {
		name          string
		slice         []string
		delete        []string
		expectedSlice []string
	}{
		{
			name:          "case 0",
			slice:         []string{"a", "b", "c"},
			delete:        []string{"b", "c"},
			expectedSlice: []string{"a"},
		},
		{
			name:          "case 1",
			slice:         []string{"a", "b", "c"},
			delete:        []string{"a", "d"},
			expectedSlice: []string{"b", "c"},
		},
		{
			name:          "case 2",
			slice:         []string{"a", "b", "c"},
			delete:        []string{"d", "c", "b", "a"},
			expectedSlice: []string{},
		},
		{
			name:          "case 3",
			slice:         []string{},
			delete:        []string{"a", "b"},
			expectedSlice: []string{},
		},
		{
			name:          "case 4",
			slice:         []string{"a", "b"},
			delete:        []string{},
			expectedSlice: []string{"a", "b"},
		},
		{
			name:          "case 5",
			slice:         []string{},
			delete:        []string{},
			expectedSlice: []string{},
		},
		{
			name:          "case 6",
			slice:         []string{"a", "b"},
			delete:        []string{"a", "b", "b", "a"},
			expectedSlice: []string{},
		},
		{
			name:          "case 7",
			slice:         []string{"a"},
			delete:        []string{"b", "b", "b", "c"},
			expectedSlice: []string{"a"},
		},
		{
			name:          "case 8",
			slice:         []string{"a", "a", "b", "b", "a", "a", "b", "b"},
			delete:        []string{"a", "a", "c", "a"},
			expectedSlice: []string{"b", "b", "b", "b"},
		},
		{
			name:          "case 9",
			slice:         []string{"a", "a", "a", "b"},
			delete:        []string{},
			expectedSlice: []string{"a", "a", "a", "b"},
		},
	}

	for _, tc := range testCases {
		slice := stringSlice(tc.slice).Delete(tc.delete...)

		if !reflect.DeepEqual(slice, tc.expectedSlice) {
			t.Fatalf("slice == %v, want %v", slice, tc.expectedSlice)
		}
	}
}
