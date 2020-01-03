// +build k8srequired

package bastion

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ghodss/yaml"
	"github.com/giantswarm/backoff"
	e2eclientsaws "github.com/giantswarm/e2eclients/aws"
	"github.com/giantswarm/e2esetup/privaterepo"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/integration/env"
)

type UserData struct {
	Ignition ignition.Ignition
	Passwd   ignition.Passwd
}

const (
	bastionIgnitionKey = "ignition.json"
)

type Config struct {
	AWSClient    *e2eclientsaws.Client
	ClusterID    string
	ImageID      string
	InstanceType string
	Logger       micrologger.Logger
}

type Bastion struct {
	// Config
	awsClient    *e2eclientsaws.Client
	clusterID    string
	imageID      string
	instanceType string
	logger       micrologger.Logger

	// VPC
	bastionSecurityGroupID *string
	workerSecurityGroupID  *string
	subnetID               *string
	vpcID                  *string

	// EC2
	instanceID *string

	// S3
	ignitionBucket *string
	ignitionURL    *string
	ignitionHash   *string
}

func NewBastion(config Config) (*Bastion, error) {
	if config.AWSClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWSClient must not be empty", config)
	}
	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.ImageID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ImageID must not be empty", config)
	}
	if config.InstanceType == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstanceType must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	return &Bastion{
		awsClient:    config.AWSClient,
		clusterID:    config.ClusterID,
		imageID:      config.ImageID,
		instanceType: config.InstanceType,
		logger:       config.Logger,
	}, nil
}

func (b *Bastion) EnsureCreated(ctx context.Context) error {
	o := func() error {
		return b.ensureCreatedOnce(ctx)
	}
	bo := backoff.NewMaxRetries(10, 1*time.Minute)
	n := backoff.NewNotifier(b.logger, ctx)
	err := backoff.RetryNotify(o, bo, n)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bastion) EnsureDeleted(ctx context.Context) error {
	o := func() error {
		return b.ensureDeletedOnce(ctx)
	}
	bo := backoff.NewMaxRetries(10, 1*time.Minute)
	n := backoff.NewNotifier(b.logger, ctx)
	err := backoff.RetryNotify(o, bo, n)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bastion) ensureCreatedOnce(ctx context.Context) error {
	err := b.ensureSecurityGroupsCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureIgnitionCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We already have an instance ID, make sure the instance exists in EC2
	if b.instanceID != nil {
		input := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				b.instanceID,
			},
		}
		response, err := b.awsClient.EC2.DescribeInstancesWithContext(ctx, input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() != "InvalidInstanceID.NotFound" {
					return microerror.Mask(err)
				}
			} else {
				return microerror.Mask(err)
			}
		}

		if len(response.Reservations) == 0 ||
			len(response.Reservations[0].Instances) == 0 {
			b.logger.LogCtx(ctx, "level", "warning", "message", "bastion instance not found", "instance_id", b.instanceID)
			b.instanceID = nil
		}
	}

	// Create the instance if it doesn't exist yet
	if b.instanceID == nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

		// User data points to full ignition in S3
		userData := UserData{
			Ignition: ignition.Ignition{
				Version: "2.2.0",
				Config: ignition.IgnitionConfig{
					Append: []ignition.ConfigReference{
						{
							Source: *b.ignitionURL,
							Verification: ignition.Verification{
								Hash: b.ignitionHash,
							},
						},
					},
				},
			},
		}
		userDataJSON, err := json.Marshal(userData)
		userDataEncoded := base64.StdEncoding.EncodeToString(userDataJSON)

		i := &ec2.RunInstancesInput{
			ImageId:      aws.String(b.imageID),
			InstanceType: aws.String(b.instanceType),
			MaxCount:     aws.Int64(1),
			MinCount:     aws.Int64(1),
			NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
				{
					AssociatePublicIpAddress: aws.Bool(true),
					DeviceIndex:              aws.Int64(0),
					Groups: []*string{
						b.bastionSecurityGroupID,
						b.workerSecurityGroupID,
					},
					SubnetId: b.subnetID,
				},
			},
			TagSpecifications: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(b.clusterID + "-bastion"),
						},
						{
							Key:   aws.String("giantswarm.io/cluster"),
							Value: aws.String(b.clusterID),
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

		_, err = b.awsClient.EC2.RunInstancesWithContext(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion instance")
	}

	return nil
}

func (b *Bastion) ensureSecurityGroupsCreated(ctx context.Context) error {
	if b.bastionSecurityGroupID != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "checking bastion security group", "group_id", b.bastionSecurityGroupID)

		i := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{
				b.bastionSecurityGroupID,
			},
		}
		o, err := b.awsClient.EC2.DescribeSecurityGroupsWithContext(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) != 1 {
			b.logger.LogCtx(ctx, "level", "warning", "message", "bastion security group not found", "group_id", b.bastionSecurityGroupID)
			b.workerSecurityGroupID = nil
		}
		b.logger.LogCtx(ctx, "level", "debug", "message", "checked bastion security group", "group_id", b.bastionSecurityGroupID)
	}

	if b.vpcID == nil || b.subnetID == nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "finding public subnet and vpc")

		i := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(b.clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("PublicSubnet")},
				},
			},
		}

		o, err := b.awsClient.EC2.DescribeSubnetsWithContext(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.Subnets) != 1 {
			return microerror.Maskf(executionFailedError, "expected one subnet, got %d", len(o.Subnets))
		}

		b.subnetID = o.Subnets[0].SubnetId
		b.vpcID = o.Subnets[0].VpcId

		b.logger.LogCtx(ctx, "level", "debug", "message", "found public subnet and vpc", "subnet_id", b.subnetID, "vpc_id", b.vpcID)
	}

	if b.workerSecurityGroupID == nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "finding worker security group")

		i := &ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:giantswarm.io/cluster"),
					Values: []*string{aws.String(b.clusterID)},
				},
				{
					Name:   aws.String("tag:aws:cloudformation:logical-id"),
					Values: []*string{aws.String("WorkerSecurityGroup")},
				},
			},
		}

		o, err := b.awsClient.EC2.DescribeSecurityGroupsWithContext(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) != 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		b.workerSecurityGroupID = o.SecurityGroups[0].GroupId

		b.logger.LogCtx(ctx, "level", "debug", "message", "found worker security group", "group_id", b.workerSecurityGroupID)
	}

	// We need to create a separate security group in order to allow SSH access to
	// the bastion instance. The AWS API does not allow tagging the security group
	// when creating it. That is why we need to separately create tags below, so
	// we are able to find it later on when we want to clean it up.
	if b.bastionSecurityGroupID == nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion security group")

		groupInput := &ec2.CreateSecurityGroupInput{
			Description: aws.String("Allow SSH access from everywhere to port 22."),
			GroupName:   aws.String(b.clusterID + "-bastion"),
			VpcId:       b.vpcID,
		}

		o, err := b.awsClient.EC2.CreateSecurityGroupWithContext(ctx, groupInput)
		if err != nil {
			return microerror.Mask(err)
		}

		tagsInput := &ec2.CreateTagsInput{
			Resources: []*string{
				b.bastionSecurityGroupID,
			},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(b.clusterID + "-bastion"),
				},
				{
					Key:   aws.String("giantswarm.io/cluster"),
					Value: aws.String(b.clusterID),
				},
			},
		}

		_, err = b.awsClient.EC2.CreateTagsWithContext(ctx, tagsInput)
		if err != nil {
			return microerror.Mask(err)
		}

		ingressInput := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: b.bastionSecurityGroupID,
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

		_, err = b.awsClient.EC2.AuthorizeSecurityGroupIngressWithContext(ctx, ingressInput)
		if err != nil {
			return microerror.Mask(err)
		}

		b.bastionSecurityGroupID = o.GroupId
		b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion security group", "group_id", b.bastionSecurityGroupID)
	}

	return nil
}

func (b *Bastion) ensureIgnitionCreated(ctx context.Context) error {
	var err error

	if b.ignitionBucket != nil {
		headBucketInput := &s3.HeadBucketInput{
			Bucket: b.ignitionBucket,
		}
		_, err := b.awsClient.S3.HeadBucketWithContext(ctx, headBucketInput)
		if err != nil {
			if err.Error() != s3.ErrCodeNoSuchBucket {
				return microerror.Mask(err)
			} else {
				b.ignitionBucket = nil
				b.ignitionURL = nil
				b.ignitionHash = nil
			}
		}
	}

	if b.ignitionBucket == nil {
		ignitionBucket := fmt.Sprintf("%s-bastion", b.clusterID)
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating S3 bucket", "bucket", ignitionBucket)

		createBucketInput := &s3.CreateBucketInput{
			Bucket: aws.String(ignitionBucket),
		}

		_, err = b.awsClient.S3.CreateBucketWithContext(ctx, createBucketInput)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "created S3 bucket", "bucket", ignitionBucket)
		b.ignitionBucket = &ignitionBucket
	}

	if b.ignitionURL == nil {
		ignitionURL := fmt.Sprintf("s3://%s/%s", b.ignitionBucket, bastionIgnitionKey)
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating S3 object", "key", bastionIgnitionKey, "bucket", b.ignitionBucket)

		userData, err := generateUserData(ctx, b.logger)
		if err != nil {
			return microerror.Mask(err)
		}

		putObjectInput := &s3.PutObjectInput{
			Key:           aws.String(bastionIgnitionKey),
			Body:          bytes.NewReader(userData),
			Bucket:        b.ignitionBucket,
			ContentLength: aws.Int64(int64(len(userData))),
		}

		_, err = b.awsClient.S3.PutObjectWithContext(ctx, putObjectInput)
		if err != nil {
			return microerror.Mask(err)
		}

		sum := sha512.Sum512(userData)
		hexSum := hex.EncodeToString(sum[:])
		ignitionHash := fmt.Sprintf("sha512-%s", hexSum)

		b.ignitionURL = &ignitionURL
		b.ignitionHash = &ignitionHash

		b.logger.LogCtx(ctx, "level", "debug", "message", "created S3 object", "key", bastionIgnitionKey, "bucket", b.ignitionBucket)
	}

	return nil
}

func (b *Bastion) ensureDeletedOnce(ctx context.Context) error {
	err := b.ensureSecurityGroupDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureIgnitionDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if b.instanceID != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "terminating bastion instance", "instance_id", b.instanceID)

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				b.instanceID,
			},
		}

		_, err = b.awsClient.EC2.TerminateInstancesWithContext(ctx, i)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() != "InvalidInstanceID.NotFound" {
					return microerror.Mask(err)
				}
			} else {
				return microerror.Mask(err)
			}
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "terminated bastion instance", "instance_id", b.instanceID)
	}

	return nil
}

func (b *Bastion) ensureSecurityGroupDeleted(ctx context.Context) error {
	if b.bastionSecurityGroupID != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "deleting bastion security group", "group_id", b.bastionSecurityGroupID)

		i := &ec2.DeleteSecurityGroupInput{
			GroupId: b.bastionSecurityGroupID,
		}
		_, err := b.awsClient.EC2.DeleteSecurityGroupWithContext(ctx, i)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidSecurityGroupID.NotFound" {
					b.bastionSecurityGroupID = nil
				} else {
					return microerror.Mask(awsErr)
				}
			}
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "deleted bastion security group", "group_id", b.bastionSecurityGroupID)
	}

	return nil
}

func (b *Bastion) ensureIgnitionDeleted(ctx context.Context) error {
	if b.ignitionURL != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "deleting ignition object", "key", bastionIgnitionKey, "bucket", b.ignitionBucket)

		i := &s3.DeleteObjectInput{
			Bucket: b.ignitionBucket,
			Key:    aws.String(bastionIgnitionKey),
		}
		_, err := b.awsClient.S3.DeleteObjectWithContext(ctx, i)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == s3.ErrCodeNoSuchKey {
					b.ignitionURL = nil
					b.ignitionHash = nil
				} else {
					return microerror.Mask(awsErr)
				}
			} else {
				return microerror.Mask(err)
			}
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "deleted ignition object", "key", bastionIgnitionKey, "bucket", b.ignitionBucket)
	}

	if b.ignitionBucket != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "deleting ignition bucket", "bucket", b.ignitionBucket)

		i := &s3.DeleteBucketInput{
			Bucket: b.ignitionBucket,
		}
		_, err := b.awsClient.S3.DeleteBucketWithContext(ctx, i)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == s3.ErrCodeNoSuchBucket {
					b.ignitionBucket = nil
				} else {
					return microerror.Mask(awsErr)
				}
			} else {
				return microerror.Mask(err)
			}
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "deleted ignition bucket", "bucket", b.ignitionBucket)
	}

	return nil
}

func generateUserData(ctx context.Context, logger micrologger.Logger) ([]byte, error) {
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

	logger.LogCtx(ctx, "level", "debug", "message", "read users from installations", "count", len(sshUserList))

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
