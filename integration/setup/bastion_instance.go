// +build k8srequired

package setup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/ghodss/yaml"
	"github.com/giantswarm/e2esetup/privaterepo"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

// TODO make the resource management more reliable to ensure proper setup and
// teardown.
//
//     https://github.com/giantswarm/giantswarm/issues/5694
//

func ensureBastionHostCreated(ctx context.Context, clusterID string, config Config) error {
	var err error

	var subnetID string
	var vpcID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "finding public subnet and vpc")

		i := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("PublicSubnet")},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeSubnets(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.Subnets) != 1 {
			return microerror.Maskf(executionFailedError, "expected one subnet, got %d", len(o.Subnets))
		}

		subnetID = *o.Subnets[0].SubnetId
		vpcID = *o.Subnets[0].VpcId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found public subnet %#q and vpc %#q", subnetID, vpcID))
	}

	var workerSecurityGroupID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "finding worker security group")

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("WorkerSecurityGroup")},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) != 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		workerSecurityGroupID = *o.SecurityGroups[0].GroupId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found worker security group %#q", workerSecurityGroupID))
	}

	// We need to create a separate security group in order to allow SSH access to
	// the bastion instance. The AWS API does not allow tagging the security group
	// when creating it. That is why we need to separately create tags below, so
	// we are able to find it later on when we want to clean it up.
	var bastionSecurityGroupID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating bastion security group")

		i := &ec2.CreateSecurityGroupInput{
			Description: aws.String("Allow SSH access from everywhere to port 22."),
			GroupName:   aws.String(clusterID + "-bastion"),
			VpcId:       aws.String(vpcID),
		}

		o, err := config.AWSClient.EC2.CreateSecurityGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}

		bastionSecurityGroupID = *o.GroupId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created bastion security group %#q", bastionSecurityGroupID))
	}

	// The AWS API does not allow tagging the security group when creating it.
	// That is why we need to separately create tags below, so we are able to find
	// it later on when we want to clean it up.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "tagging bastion security group")

		i := &ec2.CreateTagsInput{
			Resources: []*string{
				aws.String(bastionSecurityGroupID),
			},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(clusterID + "-bastion"),
				},
				{
					Key:   aws.String("giantswarm.io/cluster"),
					Value: aws.String(clusterID),
				},
			},
		}

		_, err = config.AWSClient.EC2.CreateTags(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "tagged bastion security group")
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "updating bastion security group to allow ssh access")

		i := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(bastionSecurityGroupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(22),
					IpProtocol: aws.String("tcp"),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp:      aws.String("0.0.0.0/0"),
							Description: aws.String("Allow SSH access from everywhere to port 22."),
						},
					},
					ToPort: aws.Int64(22),
				},
			},
		}

		_, err = config.AWSClient.EC2.AuthorizeSecurityGroupIngress(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "updated bastion security group to allow ssh access")
	}

	var userData string
	{
		userData, err = generateBastionUserData(ctx, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

		i := &ec2.RunInstancesInput{
			ImageId:      aws.String("ami-015e6cb33a709348e"),
			InstanceType: aws.String("t2.micro"),
			MaxCount:     aws.Int64(1),
			MinCount:     aws.Int64(1),
			NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
				{
					AssociatePublicIpAddress: aws.Bool(true),
					DeviceIndex:              aws.Int64(0),
					Groups: []*string{
						aws.String(bastionSecurityGroupID),
						aws.String(workerSecurityGroupID),
					},
					SubnetId: aws.String(subnetID),
				},
			},
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(clusterID + "-bastion"),
						},
						{
							Key:   aws.String("giantswarm.io/cluster"),
							Value: aws.String(clusterID),
						},
						{
							Key:   aws.String("giantswarm.io/instance"),
							Value: aws.String("e2e-bastion"),
						},
					},
				},
			},
			UserData: aws.String(userData),
		}

		_, err := config.AWSClient.EC2.RunInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "created bastion instance")
	}

	return nil
}

func ensureBastionHostDeleted(ctx context.Context, clusterID string, config Config) error {
	var err error

	{
		err = terminateBastionInstance(ctx, clusterID, config)
		if IsNotExists(err) {
			// This here might happen in case the bastion instance got already
			// deleted.
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = deleteBastionSecurityGroup(ctx, clusterID, config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func deleteBastionSecurityGroup(ctx context.Context, clusterID string, config Config) error {
	var err error

	var bastionSecurityGroupID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "finding bastion security group")

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []*string{aws.String(clusterID + "-bastion")},
				},
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(clusterID)},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) == 0 {
			// The security group got already deleted, so we are fine.
			return nil
		}
		if len(o.SecurityGroups) != 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		bastionSecurityGroupID = *o.SecurityGroups[0].GroupId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found bastion security group %#q", bastionSecurityGroupID))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "deleting bastion security group")

		i := &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(bastionSecurityGroupID),
		}

		_, err = config.AWSClient.EC2.DeleteSecurityGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "deleted bastion security group")
	}

	return nil
}

func terminateBastionInstance(ctx context.Context, clusterID string, config Config) error {
	var err error

	var instanceID string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "finding bastion instance id")

		i := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{
						aws.String(clusterID),
					},
				},
				{
					Name: aws.String("tag:giantswarm.io/instance"),
					Values: []*string{
						aws.String("e2e-bastion"),
					},
				},
				{
					Name: aws.String("instance-state-name"),
					Values: []*string{
						aws.String(ec2.InstanceStateNamePending),
						aws.String(ec2.InstanceStateNameRunning),
						aws.String(ec2.InstanceStateNameStopped),
						aws.String(ec2.InstanceStateNameStopping),
					},
				},
			},
		}

		o, err := config.AWSClient.EC2.DescribeInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(o.Reservations) == 0 {
			// The instance got already terminated, so we are fine.
			return nil
		}
		if len(o.Reservations) != 1 {
			return microerror.Maskf(executionFailedError, "expected one bastion instance, got %d", len(o.Reservations))
		}
		if len(o.Reservations[0].Instances) != 1 {
			return microerror.Maskf(executionFailedError, "expected one bastion instance, got %d", len(o.Reservations[0].Instances))
		}

		instanceID = *o.Reservations[0].Instances[0].InstanceId

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found bastion instance id %#q", instanceID))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "terminating bastion instance")

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceID),
			},
		}

		_, err = config.AWSClient.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "terminated bastion instance")
	}

	return nil
}

func generateBastionUserData(ctx context.Context, config Config) (string, error) {
	type User struct {
		Name              string   `json:"name"`
		Groups            []string `json:"groups"`
		SshAuthorizedKeys []string `json:"sshAuthorizedKeys"`
		Shell             string   `json:"shell"`
	}
	type Passwd struct {
		Users []User `json:"users"`
	}
	type IgnitionConfig struct {
		Version string `json:"version"`
	}
	type UserData struct {
		Ignition IgnitionConfig `json:"ignition"`
		Passwd   Passwd         `json:"passwd"`
	}

	var sshUserList []User
	var err error

	var privateRepo *privaterepo.PrivateRepo
	{
		c := privaterepo.Config{
			Owner: "giantswarm",
			Repo:  "installations",
			Token: env.GithubToken(),
		}

		privateRepo, err = privaterepo.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	{
		content, err := privateRepo.Content(ctx, "default-terraform-bastion-users.yaml")
		if err != nil {
			return "", microerror.Mask(err)
		}

		var userData UserData
		err = yaml.Unmarshal([]byte(content), &userData)
		if err != nil {
			return "", microerror.Mask(err)
		}
		sshUserList = userData.Passwd.Users
	}

	userData := UserData{
		Ignition: IgnitionConfig{
			Version: "2.1.0",
		},
		Passwd: Passwd{
			Users: sshUserList,
		},
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return base64.StdEncoding.EncodeToString(userDataJSON), nil
}
