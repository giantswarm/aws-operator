package detection

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

type Config struct {
	Logger micrologger.Logger
}

// Detection is a service implementation deciding if a tenant cluster should be
// updated or scaled.
type Detection struct {
	logger micrologger.Logger
}

func New(config Config) (*Detection, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	d := &Detection{
		logger: config.Logger,
	}

	return d, nil
}

// ShouldScale determines whether the reconciled tenant cluster should be
// scaled. A tenant cluster is only allowed to scale in the following cases.
//
//     The tenant cluster's scaling max changes.
//     The tenant cluster's scaling min changes.
//
func (d *Detection) ShouldScale(ctx context.Context, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MaxSize != key.MachineDeploymentScalingMax(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling max changes: cc.Status.TenantCluster.TCCP.ASG.MaxSize is %d while key.MachineDeploymentScalingMax(md) is %d", cc.Status.TenantCluster.TCCP.ASG.MaxSize, key.MachineDeploymentScalingMax(md)))
		return true, nil
	}
	if !cc.Status.TenantCluster.TCCP.ASG.IsEmpty() && cc.Status.TenantCluster.TCCP.ASG.MinSize != key.MachineDeploymentScalingMin(md) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should scale due to scaling min changes: cc.Status.TenantCluster.TCCP.ASG.MinSize is %d while key.MachineDeploymentScalingMin(md) is %d", cc.Status.TenantCluster.TCCP.ASG.MinSize, key.MachineDeploymentScalingMin(md)))
		return true, nil
	}

	return false, nil
}

// ShouldUpdate determines whether the reconciled tenant cluster should be
// updated. A tenant cluster is only allowed to update in the following cases.
//
//     The master node's instance type changes.
//     The worker node's docker volume size changes.
//     The worker node's instance type changes.
//     The tenant cluster's version changes.
//
func (d *Detection) ShouldUpdate(ctx context.Context, cl v1alpha1.Cluster, md v1alpha1.MachineDeployment) (bool, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if !availabilityZonesEqual(cc.Spec.TenantCluster.TCCP.AvailabilityZones, cc.Status.TenantCluster.TCCP.AvailabilityZones) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprint("detected the tenant cluster should update due to availability zone changes"))
		return true, nil
	}

	if cc.Status.TenantCluster.MasterInstance.Type != key.MasterInstanceType(cl) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to master instance type changes: cc.Status.TenantCluster.MasterInstance.Type is %q while key.MasterInstanceType(cl) is %q", cc.Status.TenantCluster.MasterInstance.Type, key.MasterInstanceType(cl)))
		return true, nil
	}
	if cc.Status.TenantCluster.VersionBundleVersion != key.OperatorVersion(&cl) {
		d.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("detected the tenant cluster should update due to version bundle version changes: cc.Status.TenantCluster.VersionBundleVersion is %q while key.OperatorVersion(&cl) is %q", cc.Status.TenantCluster.VersionBundleVersion, key.OperatorVersion(&cl)))
		return true, nil
	}

	return false, nil
}

func availabilityZonesEqual(spec []controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone, status []controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) bool {
	if spec == nil && status != nil {
		return false
	}

	if spec != nil && status == nil {
		return false
	}

	if len(spec) != len(status) {
		return false
	}

	for i, az := range spec {
		if !availabilityZoneEqual(az, status[i]) {
			return false
		}
	}

	return true
}

func availabilityZoneEqual(spec controllercontext.ContextSpecTenantClusterTCCPAvailabilityZone, status controllercontext.ContextStatusTenantClusterTCCPAvailabilityZone) bool {
	if spec.Name != status.Name {
		return false
	}

	if spec.Subnet.Private.CIDR.String() != status.Subnet.Private.CIDR.String() {
		return false
	}

	if spec.Subnet.Private.ID != status.Subnet.Private.ID {
		return false
	}

	if spec.Subnet.Public.CIDR.String() != status.Subnet.Public.CIDR.String() {
		return false
	}

	if spec.Subnet.Public.ID != status.Subnet.Public.ID {
		return false
	}

	return true
}
