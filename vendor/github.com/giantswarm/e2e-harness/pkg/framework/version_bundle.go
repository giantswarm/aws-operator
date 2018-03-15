package framework

import (
	"fmt"
	"log"
	"sort"

	"github.com/giantswarm/versionbundle"
)

func GetVersionBundleVersion(bundle []versionbundle.Bundle, vType string) (string, error) {
	validVTypes := []string{"", "current", "wip"}
	var isValid bool
	for _, v := range validVTypes {
		if v == vType {
			isValid = true
			break
		}
	}
	if !isValid {
		return "", fmt.Errorf("%q is not a valid version bundle version type", vType)
	}

	var output string
	log.Printf("Tested version %q", vType)

	// sort bundle by time to get the newest vbv.
	s := versionbundle.SortBundlesByTime(bundle)
	sort.Sort(sort.Reverse(s))
	for _, v := range s {
		if (vType == "current" || vType == "") && !v.Deprecated && !v.WIP {
			output = v.Version
			break
		}
		if vType == "wip" && v.WIP {
			output = v.Version
			break
		}
	}
	log.Printf("Version Bundle Version %q", output)
	return output, nil
}
