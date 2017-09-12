package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
)

func subnetName(cluster awstpr.CustomObject, suffix string) string {
	return fmt.Sprintf("%s-%s", cluster.Name, suffix)
}
