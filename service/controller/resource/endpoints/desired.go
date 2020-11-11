package endpoints

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var instances []*ec2.Instance
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding master instances")

		instances, err = r.searchMasterInstances(ctx, cr)
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

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d master instances", len(instances)))
	}

	var addresses []corev1.EndpointAddress
	for _, i := range instances {
		addresses = append(addresses, corev1.EndpointAddress{IP: *i.PrivateIpAddress})
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
				Addresses: addresses,
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

func (r Resource) searchMasterInstances(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) ([]*ec2.Instance, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var instances []*ec2.Instance
	{
		instancesInput := &ec2.DescribeInstancesInput{
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

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeInstances(instancesInput)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			return nil, microerror.Maskf(notFoundError, "master instance")
		}

		// check for health of the instance
		instanceHealthInput := &elb.DescribeInstanceHealthInput{
			LoadBalancerName: aws.String(key.ELBNameAPI(&cr)),
		}

		o2, err := cc.Client.TenantCluster.AWS.ELB.DescribeInstanceHealth(instanceHealthInput)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, r := range o.Reservations {
			for _, i := range r.Instances {
				for _, iState := range o2.InstanceStates {
					if *i.InstanceId == *iState.InstanceId && *iState.State == key.ELBInstanceStateInService {
						instances = append(instances, i)
					}
				}
			}

		}
	}

	if len(instances) == 0 {
		return nil, microerror.Maskf(notFoundError, "no healthy master instances found")
	}

	return instances, nil
}
