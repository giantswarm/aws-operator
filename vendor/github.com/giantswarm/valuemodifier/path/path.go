package path

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	yamltojson "github.com/ghodss/yaml"
	"github.com/giantswarm/microerror"
	"github.com/spf13/cast"
	yaml "gopkg.in/yaml.v2"
)

const (
	escapedSeparatorPlaceholder = "%%PLACEHOLDER%%"
)

var (
	placeholderExpression = regexp.MustCompile(escapedSeparatorPlaceholder)
)

// Config represents the configuration used to create a new path service.
type Config struct {
	// Settings.
	InputBytes []byte
	Separator  string
}

// DefaultConfig provides a default configuration to create a new path service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Settings.
		InputBytes: nil,
		Separator:  ".",
	}
}

// New creates a new configured path service.
func New(config Config) (*Service, error) {
	// Settings.
	if config.InputBytes == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.InputBytes must not be empty")
	}
	if config.Separator == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Separator must not be empty")
	}

	var err error

	var isJSON bool
	var jsonBytes []byte
	var jsonStructure interface{}
	{
		jsonBytes, isJSON, err = toJSON(config.InputBytes)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err := json.Unmarshal(jsonBytes, &jsonStructure)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Internals.
		isJSON:                     isJSON,
		jsonBytes:                  jsonBytes,
		jsonStructure:              jsonStructure,
		escapedSeparatorExpression: regexp.MustCompile(fmt.Sprintf(`\\%s`, config.Separator)),
		separatorExpression:        regexp.MustCompile(fmt.Sprintf(`\%s`, config.Separator)),

		// Settings.
		separator: config.Separator,
	}

	return newService, nil
}

// Service implements the path service.
type Service struct {
	// Internals.
	isJSON                     bool
	jsonBytes                  []byte
	jsonStructure              interface{}
	escapedSeparatorExpression *regexp.Regexp
	separatorExpression        *regexp.Regexp

	// Settings.
	separator string
}

// All returns all paths found in the configured JSON structure.
func (s *Service) All() ([]string, error) {
	paths, err := s.allFromInterface(s.jsonStructure)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	sort.Strings(paths)

	return paths, nil
}

// Get returns the value found under the given path, if any.
func (s *Service) Get(path string) (interface{}, error) {
	value, err := s.getFromInterface(s.escapeKey(path), s.jsonStructure)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return value, nil
}

func (s *Service) OutputBytes() ([]byte, error) {
	b := s.jsonBytes
	if !s.isJSON {
		var err error
		b, err = yamltojson.JSONToYAML(b)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return b, nil
}

// Set changes the value of the given path.
func (s *Service) Set(path string, value interface{}) error {
	var err error

	s.jsonStructure, err = s.setFromInterface(s.escapeKey(path), value, s.jsonStructure)
	if err != nil {
		return microerror.Mask(err)
	}

	b, err := json.MarshalIndent(s.jsonStructure, "", "  ")
	if err != nil {
		return microerror.Mask(err)
	}
	s.jsonBytes = b

	return nil
}

func (s *Service) Validate(paths []string) error {
	all, err := s.All()
	if err != nil {
		return microerror.Mask(err)
	}

	var trimmedAll []string
	for _, service := range all {
		pv := strings.Split(service, ".")
		trimmedAll = append(trimmedAll, pv[len(pv)-1])
	}

	for _, p := range paths {
		fields := trimmedAll
		if strings.Index(p, ".") != -1 {
			fields = all
		}
		if containsString(fields, p) {
			continue
		}

		return microerror.Maskf(notFoundError, "path '%s'", p)
	}

	return nil
}

func (s *Service) allFromInterface(value interface{}) ([]string, error) {
	// process map
	{
		stringMap, err := cast.ToStringMapE(value)
		if err != nil {
			// fall through
		} else {
			var paths []string

			for k, v := range stringMap {
				ps, err := s.allFromInterface(v)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				k := s.separatorExpression.ReplaceAllString(k, fmt.Sprintf(`\%s`, s.separator))

				if ps != nil {
					for _, p := range ps {
						paths = append(paths, fmt.Sprintf("%s%s%s", k, s.separator, p))
					}
				} else {
					paths = append(paths, k)
				}
			}

			return paths, nil
		}
	}

	// process slice
	{
		slice, err := cast.ToSliceE(value)
		if err != nil {
			// fall through
		} else {
			var paths []string

			for i, v := range slice {
				ps, err := s.allFromInterface(v)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				for _, p := range ps {
					paths = append(paths, fmt.Sprintf("[%d]%s%s", i, s.separator, p))
				}
			}

			return paths, nil
		}
	}

	// process string
	{
		str, err := cast.ToStringE(value)
		if err != nil {
			// fall through
		} else if str == "" {
			// fall through
		} else {
			jsonBytes, _, err := toJSON([]byte(str))
			if err != nil {
				// fall through
			} else {
				var jsonStructure interface{}
				err := json.Unmarshal(jsonBytes, &jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				ps, err := s.allFromInterface(jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				return ps, nil
			}
		}
	}

	return nil, nil
}

func (s *Service) escapeKey(key string) string {
	return s.escapedSeparatorExpression.ReplaceAllString(key, escapedSeparatorPlaceholder)
}

func (s *Service) getFromInterface(path string, jsonStructure interface{}) (interface{}, error) {
	split := strings.Split(path, s.separator)
	key := s.unescapeKey(split[0])

	// process map
	{
		stringMap, err := cast.ToStringMapE(jsonStructure)
		if err != nil {
			// fall through
		} else {
			value, ok := stringMap[key]
			if ok {
				if len(split) == 1 {
					return value, nil
				} else {
					recPath := strings.Join(split[1:], s.separator)

					v, err := s.getFromInterface(recPath, value)
					if err != nil {
						return nil, microerror.Mask(err)
					}

					return v, nil
				}
			} else {
				return nil, microerror.Maskf(notFoundError, "key '%s'", path)
			}
		}
	}

	// process slice
	{
		slice, err := cast.ToSliceE(jsonStructure)
		if err != nil {
			// fall through
		} else {
			index, err := indexFromKey(key)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if index >= len(slice) {
				return nil, microerror.Maskf(notFoundError, "key '%s'", key)
			}
			recPath := strings.Join(split[1:], s.separator)

			v, err := s.getFromInterface(recPath, slice[index])
			if err != nil {
				return nil, microerror.Mask(err)
			}

			return v, nil
		}
	}

	// process string
	{
		str, err := cast.ToStringE(jsonStructure)
		if err != nil {
			// fall through
		} else {
			jsonBytes, _, err := toJSON([]byte(str))
			if err != nil {
				// fall through
			} else {
				var jsonStructure interface{}
				err := json.Unmarshal(jsonBytes, &jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				v, err := s.getFromInterface(path, jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				return v, nil
			}
		}
	}

	return nil, nil
}

func (s *Service) setFromInterface(path string, value interface{}, jsonStructure interface{}) (interface{}, error) {
	split := strings.Split(path, s.separator)
	key := s.unescapeKey(split[0])

	// process map
	{
		_, ok := jsonStructure.(string)
		if ok {
			// Fall through in case our received JSON structure is actually a string.
			// cast.ToStringMapE was working as expected until
			// https://github.com/spf13/cast/pull/59, so we have to make sure we do
			// not call cast.ToStringMapE only if we do not have an actual string,
			// because cast.ToStringMapE would now accept the string instead of
			// returning an error like it did before.
		} else {
			stringMap, err := cast.ToStringMapE(jsonStructure)
			if err != nil {
				// fall through
			} else {
				if len(split) == 1 {
					_, ok := stringMap[path]
					if ok {
						stringMap[path] = value
						return stringMap, nil
					} else {
						return nil, microerror.Maskf(notFoundError, "key '%s'", path)
					}
				} else {
					_, ok := stringMap[key]
					if ok {
						recPath := strings.Join(split[1:], s.separator)

						modified, err := s.setFromInterface(recPath, value, stringMap[key])
						if err != nil {
							return nil, microerror.Mask(err)
						}
						stringMap[key] = modified

						return stringMap, nil
					} else {
						return nil, microerror.Maskf(notFoundError, "key '%s'", path)
					}
				}
			}
		}
	}

	// process slice
	{
		slice, err := cast.ToSliceE(jsonStructure)
		if err != nil {
			// fall through
		} else {
			index, err := indexFromKey(key)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if index >= len(slice) {
				return nil, microerror.Maskf(notFoundError, "key '%s'", key)
			}
			recPath := strings.Join(split[1:], s.separator)

			modified, err := s.setFromInterface(recPath, value, slice[index])
			if err != nil {
				return nil, microerror.Mask(err)
			}
			slice[index] = modified

			return slice, nil
		}
	}

	// process string
	{
		str, err := cast.ToStringE(jsonStructure)
		if err != nil {
			// fall through
		} else {
			jsonBytes, isJSON, err := toJSON([]byte(str))
			if err != nil {
				// fall through
			} else {
				var jsonStructure interface{}
				err := json.Unmarshal(jsonBytes, &jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				modified, err := s.setFromInterface(path, value, jsonStructure)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				var b []byte
				if !isJSON {
					b, err = yamltojson.Marshal(modified)
					if err != nil {
						return nil, microerror.Mask(err)
					}
				} else {
					b, err = json.MarshalIndent(modified, "", "  ")
					if err != nil {
						return nil, microerror.Mask(err)
					}
				}

				return string(b), nil
			}
		}
	}

	return nil, nil
}

func (s *Service) unescapeKey(key string) string {
	return placeholderExpression.ReplaceAllString(key, s.separator)
}

func containsString(list []string, item string) bool {
	for _, l := range list {
		if l == item {
			return true
		}
	}

	return false
}

func indexFromKey(key string) (int, error) {
	re := regexp.MustCompile("\\[[0-9]+\\]")
	ok := re.MatchString(key)
	if !ok {
		return 0, microerror.Maskf(keyNotIndexError, key)
	}

	s := key[1 : len(key)-1]
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return i, nil
}

func isJSON(b []byte) bool {
	var l []interface{}
	isList := json.Unmarshal(b, &l) == nil

	var m map[string]interface{}
	isObject := json.Unmarshal(b, &m) == nil

	return isObject || isList
}

func isYAMLList(b []byte) bool {
	var l []interface{}
	return yaml.Unmarshal(b, &l) == nil && bytes.HasPrefix(b, []byte("-"))
}

func isYAMLObject(b []byte) bool {
	var m map[interface{}]interface{}
	return yaml.Unmarshal(b, &m) == nil && !bytes.HasPrefix(b, []byte("-"))
}

func pathWithKey(key string, path string, separator string) string {
	return strings.Join([]string{key, path}, separator)
}

func toJSON(b []byte) ([]byte, bool, error) {
	if isJSON(b) {
		return b, true, nil
	}

	isYAMLList := isYAMLList(b)
	isYAMLObject := isYAMLObject(b)

	var jsonBytes []byte
	if isYAMLList && !isYAMLObject {
		var jsonList []interface{}
		err := yamltojson.Unmarshal(b, &jsonList)
		if err != nil {
			return nil, false, microerror.Mask(err)
		}

		jsonBytes, err = json.Marshal(jsonList)
		if err != nil {
			return nil, false, microerror.Mask(err)
		}

		return jsonBytes, false, nil
	}

	if !isYAMLList && isYAMLObject {
		var jsonMap map[string]interface{}
		err := yamltojson.Unmarshal(b, &jsonMap)
		if err != nil {
			return nil, false, microerror.Mask(err)
		}

		jsonBytes, err = json.Marshal(jsonMap)
		if err != nil {
			return nil, false, microerror.Mask(err)
		}

		return jsonBytes, false, nil
	}

	return nil, false, microerror.Mask(invalidFormatError)
}
