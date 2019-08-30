package template

import (
	"testing"
)

func Test_Template_Render(t *testing.T) {
	t.Parallel()
	tpl := "some string <{{.Value}}> another string"
	d := struct {
		Value string
	}{"myvalue"}
	expected := "some string <myvalue> another string"

	actual, err := Render([]string{tpl}, d)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if actual != expected {
		t.Errorf("unexpected output, want %q, got %q", expected, actual)
	}
}
