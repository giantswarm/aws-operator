package versionbundle

import "strconv"

func isPositiveNumber(s string) bool {
	i, err := strconv.Atoi(s)
	if err != nil {
		return false
	}

	if i < 0 {
		return false
	}

	return true
}
