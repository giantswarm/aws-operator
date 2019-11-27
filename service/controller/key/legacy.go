package key

import (
	"github.com/giantswarm/aws-operator/service/controller/internal/templates/cloudconfig"
)

// NOTE that code below is deprecated and needs refactoring.

func CloudConfigSmallTemplates() []string {
	return []string{
		cloudconfig.Small,
	}
}
