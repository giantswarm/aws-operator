package validator

import (
	"encoding/json"

	"github.com/giantswarm/microerror"
)

// StructToMap is a helper method to convert an expected request data structure
// in the correctly formatted type to UnknownAttributes.
func StructToMap(s interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return m, nil
}
