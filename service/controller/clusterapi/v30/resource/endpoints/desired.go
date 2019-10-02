package endpoints

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var instance *ec2.Instance
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding master instance")

		instance, err = r.searchMasterInstance(ctx, cr)
		if IsNotFound(err) {
			// During updates the master instance is shut down and thus cannot be found.
			// In such cases we cancel the reconciliation for the endpoint resource.
			// This should be ok since all endpoints should be created and up to date
			// already. In case we miss an update it will be done on the next resync
			// period once the master instance is up again.
			//
			// TODO we might want to alert at some point when the master instance was
			// not seen for too long. Like we should be able to find it again after
			// three resync periods max or something.
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find master instance")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found master instance")
	}

	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterEndpointsName,
			Namespace: key.ClusterID(&cr),
			Labels: map[string]string{
				"app":      masterEndpointsName,
				"cluster":  key.ClusterID(&cr),
				"customer": key.OrganizationID(&cr),
			},
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: *instance.PrivateIpAddress,
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Port: httpsPort,
					},
				},
			},
		},
	}

	return endpoints, nil
}

func (r Resource) searchMasterInstance(ctx context.Context, cr v1alpha1.Cluster) (*ec2.Instance, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var instance *ec2.Instance
	{
		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(key.MasterInstanceName(cr)),
					},
				},
				{
					Name: aws.String(fmt.Sprintf("tag:%s", key.TagCluster)),
					Values: []*string{
						aws.String(key.ClusterID(&cr)),
					},
				},
				{
					Name: aws.String("instance-state-name"),
					Values: []*string{
						aws.String(ec2.InstanceStateNameRunning),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			return nil, microerror.Maskf(notFoundError, "master instance")
		}
		if len(o.Reservations) != 1 {
			return nil, microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return nil, microerror.Maskf(executionFailedError, "expected one master instance, got %d", len(o.Reservations[0].Instances))
		}

		instance = o.Reservations[0].Instances[0]
	}

	return instance, nil
}
