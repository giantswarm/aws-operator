package helmclient

import (
	"fmt"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
)

// MergeValues merges config values so they can be used when installing or
// updating Helm releases. It takes in 2 maps with a string key and YAML values
// passed as a byte array.
//
// A deep merge is performed into a single map[string]interface{} output. If a
// value is present in both then the source map is preferred.
//
// The YAML values are parsed using yamlToStringMap. This is because the
// default behaviour of the YAML parser is to unmarshal into
// map[interface{}]interface{} which causes problems with the merge logic.
// See https://github.com/go-yaml/yaml/issues/139.
//
func MergeValues(destMap, srcMap map[string][]byte) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	destVals, err := processYAML(destMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	srcVals, err := processYAML(srcMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = mergeValues(destVals, srcVals)

	return result, nil
}

// processYAML accepts a map with a string key and YAML values passed as a
// byte array. Only a single key is supported and the input data structure
// matches the configmap or secret where the data is stored.
func processYAML(inputMap map[string][]byte) (map[string]interface{}, error) {
	var err error

	result := map[string]interface{}{}

	if len(inputMap) > 1 {
		return nil, microerror.Maskf(executionFailedError, "merging %d keys is unsupported expected 1 key", len(inputMap))
	}

	for _, v := range inputMap {
		result, err = yamlToStringMap(v)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return result, nil
}

// mergeValues implements the merge logic. It performs a deep merge. If a value
// is present in both then the source map is preferred.
//
// Logic is based on the upstream logic implemented by Helm.
// https://github.com/helm/helm/blob/240e539cec44e2b746b3541529d41f4ba01e77df/cmd/helm/install.go#L358
func mergeValues(dest, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		if _, exists := dest[k]; !exists {
			// If the key doesn't exist already. Set the key to that value.
			dest[k] = v
			continue
		}

		nextMap, ok := v.(map[string]interface{})
		if !ok {
			// If it isn't another map. Overwrite the value.
			dest[k] = v
			continue
		}

		// Edge case: If the key exists in the destination but isn't a map.
		destMap, ok := dest[k].(map[string]interface{})
		if !ok {
			// If the source map has a map for this key. Prefer that value.
			dest[k] = v
			continue
		}

		// If we got to this point. It is a map in both so merge them.
		dest[k] = mergeValues(destMap, nextMap)
	}

	return dest
}

// yamlToStringMap unmarshals the YAML input into a map[string]interface{}
// with string keys. This is necessary because the default behaviour of the
// YAML parser is to return map[interface{}]interface{} types.
// See https://github.com/go-yaml/yaml/issues/139.
//
func yamlToStringMap(input []byte) (map[string]interface{}, error) {
	var raw interface{}
	var result map[string]interface{}

	err := yaml.Unmarshal(input, &raw)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if raw == nil {
		result := map[string]interface{}{}
		return result, nil
	}

	inputMap, ok := raw.(map[interface{}]interface{})
	if !ok {
		return nil, microerror.Maskf(executionFailedError, "input type %T but expected %T", raw, inputMap)
	}

	result = processInterfaceMap(inputMap)

	return result, nil
}

func processInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = processValue(v)
	}
	return res
}

func processInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = processValue(v)
	}
	return res
}

func processValue(v interface{}) interface{} {
	// yaml null-valued key unmarshalls to nil value in Go, so nil is a valid value that has to be handled
	// See https://helm.sh/docs/chart_template_guide/#deleting-a-default-key.
	if v == nil {
		return v
	}

	switch v := v.(type) {
	case bool:
		return v
	case float64:
		return v
	case int:
		return v
	case string:
		return v
	case []interface{}:
		return processInterfaceArray(v)
	case map[interface{}]interface{}:
		return processInterfaceMap(v)
	default:
		return microerror.Maskf(executionFailedError, "value %#v with type %T not supported", v, v)
	}
}
