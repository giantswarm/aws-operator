// +build k8srequired

package setup

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ghodss/yaml"
	"github.com/giantswarm/e2esetup/privaterepo"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/integration/env"
)

type UserData struct {
	Ignition ignition.Ignition
	Passwd   ignition.Passwd
}

const (
	bastionIgnitionKey = "ignition.json"
)

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

	var bastionIgnitionBucket string
	{
		bastionIgnitionBucket = fmt.Sprintf("%s-bastion", clusterID)
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 bucket %#q", bastionIgnitionBucket))

		createBucketInput := &s3.CreateBucketInput{
			Bucket: aws.String(bastionIgnitionBucket),
		}

		_, err = config.AWSClient.S3.CreateBucket(createBucketInput)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 bucket %#q", bastionIgnitionBucket))
	}

	var bastionIgnitionObjectURL string
	var bastionIgnitionHash string

	{
		userData, err := generateBastionUserData(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		bastionIgnitionObjectURL = fmt.Sprintf("s3://%s/%s", bastionIgnitionBucket, bastionIgnitionKey)
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object %#q", bastionIgnitionObjectURL))

		putObjectInput := &s3.PutObjectInput{
			Key:           aws.String(bastionIgnitionKey),
			Body:          bytes.NewReader(userData),
			Bucket:        aws.String(bastionIgnitionBucket),
			ContentLength: aws.Int64(int64(len(userData))),
		}

		_, err = config.AWSClient.S3.PutObject(putObjectInput)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 object %#q", bastionIgnitionObjectURL))

		sum := sha512.Sum512(userData)
		bastionIgnitionHash = fmt.Sprintf("sha512-%s", hex.EncodeToString(sum[:]))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

		userData := UserData{
			Ignition: ignition.Ignition{
				Version: "2.2.0",
				Config: ignition.IgnitionConfig{
					Append: []ignition.ConfigReference{
						{
							Source: bastionIgnitionObjectURL,
							Verification: ignition.Verification{
								Hash: aws.String(bastionIgnitionHash),
							},
						},
					},
				},
			},
		}
		userDataJSON, err := json.Marshal(userData)
		userDataEncoded := base64.StdEncoding.EncodeToString(userDataJSON)

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
			UserData: aws.String(userDataEncoded),
		}

		_, err = config.AWSClient.EC2.RunInstances(i)
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

func generateBastionUserData(ctx context.Context) ([]byte, error) {
	var sshUserList []ignition.PasswdUser
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
			return nil, microerror.Mask(err)
		}
	}

	{
		content, err := privateRepo.Content(ctx, "default-terraform-bastion-users.yaml")
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var userData UserData
		err = yaml.Unmarshal([]byte(content), &userData)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		sshUserList = userData.Passwd.Users
	}

	userData := UserData{
		Ignition: ignition.Ignition{
			Version: "2.1.0",
		},
		Passwd: ignition.Passwd{
			Users: sshUserList,
		},
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return userDataJSON, nil
}
