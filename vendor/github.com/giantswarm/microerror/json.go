package microerror

import (
	"encoding/json"
	"errors"
	"fmt"
)

// JSON prints the error with enriched information in JSON format. Enriched
// information includes:
//
//	- All fields from Error type.
//	- Error stack.
//
// The rendered JSON can be unmarshalled with JSONError type.
func JSON(err error) string {
	if err == nil {
		err = &annotatedError{
			annotation: fmt.Sprintf("%v", nil),
			underlying: &Error{
				Kind: kindNil,
			},
		}
	}

	var eerr *Error
	var serr *stackedError
	if !errors.As(err, &eerr) && !errors.As(err, &serr) {
		err = &annotatedError{
			annotation: err.Error(),
			underlying: &Error{
				Kind: kindUnknown,
			},
		}
	}

	bytes, err := json.Marshal(err)
	if err != nil {
		panic(err.Error())
	}

	return string(bytes)
}
