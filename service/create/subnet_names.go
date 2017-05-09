package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func subnetName(cluster awstpr.CustomObject, prefix string) string {
	return fmt.Sprintf("%s-%s", cluster.Name, prefix)
}
