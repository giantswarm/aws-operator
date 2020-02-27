package microerror

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	kindNil     = "nil"
	kindUnknown = "unknown"
)

type Error struct {
	Desc string `json:"desc,omitempty"`
	Docs string `json:"docs,omitempty"`
	Kind string `json:"kind"`
}

// GoString is here for backward compatibility.
//
// NOTE: Use JSON(err) instead of Printf("%#v", err).
func (e *Error) GoString() string {
	return JSON(e)
}

func (e *Error) Error() string {
	return toStringCase(e.Kind)
}

type JSONError struct {
	*Error `json:",inline"`

	Annotation string       `json:"annotation,omitempty"`
	Stack      []StackEntry `json:"stack,omitempty"`
}

type StackEntry struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type annotatedError struct {
	annotation string
	underlying *Error
}

// GoString is here for backward compatibility.
//
// NOTE: Use JSON(err) instead of Printf("%#v", err).
func (e *annotatedError) GoString() string {
	return JSON(e)
}

func (e *annotatedError) Error() string {
	if e.annotation == "" {
		return e.underlying.Error()
	}
	return e.underlying.Error() + ": " + e.annotation
}

func (e *annotatedError) MarshalJSON() ([]byte, error) {
	o := JSONError{
		Error: e.underlying,

		Annotation: e.annotation,
	}

	bytes, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("microerror.annotatedError.MarshalJSON: %w object=%#v", err, o)
	}

	return bytes, nil
}

func (e *annotatedError) Unwrap() error {
	return e.underlying
}

type stackedError struct {
	stackEntry StackEntry
	underlying error
}

// GoString is here for backward compatibility.
//
// NOTE: Use JSON(err) instead of Printf("%#v", err).
func (e *stackedError) GoString() string {
	return JSON(e)
}

func (e *stackedError) Error() string {
	return e.underlying.Error()
}

// MarshalJSON unwraps all the stackedError and reconstructs the stack. Then it
// tries to find annotatedError to find the custom error annotation and finally
// tries to find Error or just creates one from arbitrary error. Then it sets
// all the fields to JSONError and finally marshals it using standard
// json.Marshal call.
func (e *stackedError) MarshalJSON() ([]byte, error) {
	var stack = []StackEntry{
		e.stackEntry,
	}
	{
		underlying := e.underlying
		var serr *stackedError
		for errors.As(underlying, &serr) {
			stack = append([]StackEntry{serr.stackEntry}, stack...)
			underlying = serr.underlying
		}
	}

	var eerr *Error
	var annotation string
	{
		if errors.As(e, &eerr) {
			var aerr *annotatedError
			if errors.As(e, &aerr) {
				annotation = aerr.annotation
			}
		} else {
			eerr = &Error{
				Kind: kindUnknown,
			}

			annotation = e.Error()
		}
	}

	o := JSONError{
		Error: eerr,

		Annotation: annotation,
		Stack:      stack,
	}

	bytes, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("microerror.stackedError.MarshalJSON: %w object=%#v", err, o)
	}

	return bytes, nil
}

func (e *stackedError) Unwrap() error {
	return e.underlying
}
