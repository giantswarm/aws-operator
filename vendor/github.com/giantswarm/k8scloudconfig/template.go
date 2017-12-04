package cloudconfig

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/k8scloudconfig/v_1_0_0"
	"github.com/giantswarm/k8scloudconfig/v_1_1_0"
	"github.com/giantswarm/k8scloudconfig/v_2_0_0"
)

type version string

const (
	V_0_1_0 version = "v_0_1_0"
	// V_1_0_0 is only used for KVM and has the experimental encryption feature
	// disabled. See also https://github.com/giantswarm/k8scloudconfig/pull/257.
	V_1_0_0 version = "v_1_0_0"
	// V_1_1_0 uses generated client types for templating but does not have the
	// experimental encryption feature enabled. We use it e.g. for KVM only since
	// the TPR migration. See also
	// https://github.com/giantswarm/k8scloudconfig/pull/259.
	V_1_1_0 version = "v_1_1_0"
	// V_2_0_0 uses generated client types for templating. We use it e.g. since
	// the TPR migration. See also
	// https://github.com/giantswarm/k8scloudconfig/pull/255.
	V_2_0_0 version = "v_2_0_0"
)

type Template struct {
	Master string
	Worker string
}

// NewTemplate returns a template structure containing cloud config templates
// versioned as provided.
//
// NOTE that version is a private type to this package to prevent users from
// accidentially providing wrong versions.
func NewTemplate(v version) (Template, error) {
	var template Template

	switch v {
	case V_0_1_0:
		template.Master = v_0_1_0.MasterTemplate
		template.Worker = v_0_1_0.WorkerTemplate
	case V_1_0_0:
		template.Master = v_1_0_0.MasterTemplate
		template.Worker = v_1_0_0.WorkerTemplate
	case V_1_1_0:
		template.Master = v_1_1_0.MasterTemplate
		template.Worker = v_1_1_0.WorkerTemplate
	case V_2_0_0:
		template.Master = v_2_0_0.MasterTemplate
		template.Worker = v_2_0_0.WorkerTemplate
	default:
		return Template{}, microerror.Maskf(notFoundError, "template version '%s'", v)
	}

	return template, nil
}
