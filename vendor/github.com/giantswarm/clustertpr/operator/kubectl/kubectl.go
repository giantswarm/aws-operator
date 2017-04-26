package kubectl

import (
	"github.com/giantswarm/clustertpr/operator/kubectl/googleapi"
)

type Kubectl struct {
	GoogleAPI googleapi.GoogleAPI
}
