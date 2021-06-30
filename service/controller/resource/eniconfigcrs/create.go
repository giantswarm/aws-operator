package eniconfigcrs

import (
	"context"

	"github.com/aws/amazon-vpc-cni-k8s/pkg/apis/crd/v1alpha1"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.Debugf(ctx, "kubernetes clients are not available in controller context yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	// We need to configure our security group ID to all the ENIConfig CRs.
	// Therefore we look it up from the controller context, where it is already
	// available.
	var securityGroupID string
	{
		securityGroupID = idFromGroups(cc.Status.TenantCluster.TCCP.SecurityGroups, key.SecurityGroupName(&cr, "aws-cni"))
	}

	// Now we compute the desired state of the ENIConfig CRs, so we can apply them
	// to the Tenant Cluster.
	var eniConfigs []*v1alpha1.ENIConfig
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		ec := &v1alpha1.ENIConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
				Kind:       "ENIConfig",
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annotation.Docs: "https://godoc.org/github.com/aws/amazon-vpc-cni-k8s/pkg/apis/crd/v1alpha1#ENIConfig",
				},
				Name:      az.Name,
				Namespace: corev1.NamespaceDefault,
			},
			Spec: v1alpha1.ENIConfigSpec{
				SecurityGroups: []string{
					securityGroupID,
				},
				Subnet: az.Subnet.AWSCNI.ID,
			},
		}

		eniConfigs = append(eniConfigs, ec)
	}

	// Here we create the ENIConfig CRs if they do not exist. If they are already
	// present, we update them according to our desired state. This is some simple
	// fire and forget approach for now.
	for _, ec := range eniConfigs {
		r.logger.Debugf(ctx, "ensuring ENIConfig CR")

		err := cc.Client.TenantCluster.K8s.CtrlClient().Create(ctx, ec)
		if errors.IsAlreadyExists(err) {
			var latest v1alpha1.ENIConfig

			err := cc.Client.TenantCluster.K8s.CtrlClient().Get(ctx, types.NamespacedName{Name: ec.GetName(), Namespace: ec.GetNamespace()}, &latest)
			if err != nil {
				return microerror.Mask(err)
			}

			ec.ResourceVersion = latest.GetResourceVersion()

			err = cc.Client.TenantCluster.K8s.CtrlClient().Update(ctx, ec)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ensured ENIConfig CR")
	}

	return nil
}

func idFromGroups(groups []*ec2.SecurityGroup, name string) string {
	for _, g := range groups {
		if awstags.ValueForKey(g.Tags, "Name") == name {
			return *g.GroupId
		}
	}

	return ""
}
