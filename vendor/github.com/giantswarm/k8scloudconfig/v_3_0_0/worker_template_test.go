package v_3_0_0

import (
	"testing"
	"text/template"
)

func Test_WorkerTemplate(t *testing.T) {
	tmpl, err := template.New("").Parse(WorkerTemplate)
	if err != nil {
		t.Fatalf("expected err = nil, got %v", err)
	}

	// Test with empty struct parameters. This should fail with field
	// evaluation.
	{
		params := struct{}{}

		err := tmpl.Execute(nopWriter{}, params)
		if err == nil {
			t.Fatalf("expected err != nil, got nil")
		}
	}

	// Test with Params struct. This should contain all evaluated fields.
	{
		params := Params{
			// Extension has to be set because it's interface and
			// template evaluation will panic otherwise.
			Extension: nopExtension{},
		}

		err := tmpl.Execute(new(nopWriter), params)
		if err != nil {
			t.Fatalf("expected err = nil, got %v", err)
		}
	}
}
