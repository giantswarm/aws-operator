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
	"github.com/giantswarm/aws-operator/integration/key"
)

type UserData struct {
	Ignition ignition.Ignition
	Passwd   ignition.Passwd
}

type Config struct {
	AccountID    string
	AWSClient    *e2eclientsaws.Client
	ClusterID    string
	ImageID      string
	InstanceType string
	Logger       micrologger.Logger
}

type Bastion struct {
	// Config
	awsClient    *e2eclientsaws.Client
	accountID    string
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
	ignitionURL  *string
	ignitionHash *string
}

func NewBastion(config Config) (*Bastion, error) {
	if config.AccountID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AccountID must not be empty", config)
	}
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

func retry(ctx context.Context, logger micrologger.Logger, operation backoff.Operation) error {
	backOff := backoff.NewMaxRetries(10, 1*time.Minute)
	notifier := backoff.NewNotifier(logger, ctx)
	err := backoff.RetryNotify(operation, backOff, notifier)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (b *Bastion) EnsureCreated(ctx context.Context) error {
	operation := func() error {
		return b.ensureCreatedOnce(ctx)
	}
	err := retry(ctx, b.logger, operation)
	if err != nil {
		return microerror.Mask(err)
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

	err = b.ensureInstanceCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (b *Bastion) ensureInstanceCreated(ctx context.Context) error {
	if b.instanceID != nil {
		return nil
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

	if b.ignitionURL == nil ||
		b.ignitionHash == nil {
		return microerror.Maskf(executionFailedError, "missing instance userdata dependencies")
	}
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

	if b.bastionSecurityGroupID == nil ||
		b.workerSecurityGroupID == nil ||
		b.subnetID == nil {
		return microerror.Maskf(executionFailedError, "missing instance network dependencies")
	}
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

	response, err := b.awsClient.EC2.RunInstancesWithContext(ctx, i)
	if err != nil {
		return microerror.Mask(err)
	}

	b.instanceID = response.Instances[0].InstanceId
	b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion instance", "instance_id", b.instanceID)

	return nil
}

func (b *Bastion) ensureSecurityGroupsCreated(ctx context.Context) error {
	if b.bastionSecurityGroupID != nil {
		return nil
	}

	// Find worker security group
	{
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

	// Create bastion security group for SSH
	var groupID string
	{
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

		groupID = *o.GroupId
	}

	// Add tags to bastion security group
	{
		tagsInput := &ec2.CreateTagsInput{
			Resources: []*string{
				aws.String(groupID),
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

		_, err := b.awsClient.EC2.CreateTagsWithContext(ctx, tagsInput)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Add SSH ingress to bastion security group
	{
		ingressInput := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(groupID),
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

		_, err := b.awsClient.EC2.AuthorizeSecurityGroupIngressWithContext(ctx, ingressInput)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	b.bastionSecurityGroupID = &groupID
	b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion security group", "group_id", b.bastionSecurityGroupID)

	return nil
}

func (b *Bastion) ensureIgnitionCreated(ctx context.Context) error {
	if b.ignitionURL != nil {
		return nil
	}

	bucketExists := true
	ignitionBucket := key.BastionIgnitionBucket(b.accountID)

	{
		headBucketInput := &s3.HeadBucketInput{
			Bucket: aws.String(ignitionBucket),
		}
		_, err := b.awsClient.S3.HeadBucketWithContext(ctx, headBucketInput)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NotFound" {
					bucketExists = false
				} else {
					return microerror.Mask(err)
				}
			} else {
				return microerror.Mask(err)
			}
		}
	}

	if !bucketExists {
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion ignition bucket", "bucket", ignitionBucket)
		createBucketInput := &s3.CreateBucketInput{
			Bucket: aws.String(ignitionBucket),
		}
		_, err := b.awsClient.S3.CreateBucketWithContext(ctx, createBucketInput)
		if err != nil {
			return microerror.Mask(err)
		}
		b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion ignition bucket", "bucket", ignitionBucket)
	}

	{
		objectKey := key.BastionIgnitionObject(b.clusterID)
		ignitionURL := key.BastionIgnitionURL(b.accountID, b.clusterID)
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion ignition object", "key", objectKey, "bucket", ignitionBucket)

		userData, err := generateUserData(ctx, b.logger)
		if err != nil {
			return microerror.Mask(err)
		}

		putObjectInput := &s3.PutObjectInput{
			Key:           aws.String(objectKey),
			Body:          bytes.NewReader(userData),
			Bucket:        aws.String(ignitionBucket),
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

		b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion ignition object", "key", objectKey, "bucket", ignitionBucket)
	}

	return nil
}

func (b *Bastion) EnsureDeleted(ctx context.Context) error {
	operation := func() error {
		return b.ensureDeletedOnce(ctx)
	}
	err := retry(ctx, b.logger, operation)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (b *Bastion) ensureDeletedOnce(ctx context.Context) error {
	err := b.ensureInstanceDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureIgnitionDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureSecurityGroupDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (b *Bastion) ensureInstanceDeleted(ctx context.Context) error {
	if b.instanceID == nil {
		return nil
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "ensuring bastion instance deleted", "instance_id", b.instanceID)

	i := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			b.instanceID,
		},
	}
	_, err := b.awsClient.EC2.TerminateInstancesWithContext(ctx, i)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidInstanceID.NotFound" {
				b.logger.LogCtx(ctx, "level", "warning", "message", "bastion instance not found", "instance_id", b.instanceID)
				b.instanceID = nil
			} else {
				return microerror.Mask(awsErr)
			}
		} else {
			return microerror.Mask(err)
		}
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "ensured bastion instance deleted", "instance_id", b.instanceID)

	return nil
}

func (b *Bastion) ensureIgnitionDeleted(ctx context.Context) error {
	if b.ignitionURL == nil {
		return nil
	}

	ignitionBucket := key.BastionIgnitionBucket(b.accountID)
	objectKey := key.BastionIgnitionObject(b.clusterID)
	b.logger.LogCtx(ctx, "level", "debug", "message", "ensuring baston ignition object deleted", "key", objectKey, "bucket", ignitionBucket)

	i := &s3.DeleteObjectInput{
		Bucket: aws.String(ignitionBucket),
		Key:    aws.String(objectKey),
	}
	_, err := b.awsClient.S3.DeleteObjectWithContext(ctx, i)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				b.logger.LogCtx(ctx, "level", "warning", "message", "bastion ignition object not found", "key", objectKey)
				b.ignitionURL = nil
				b.ignitionHash = nil
			} else {
				return microerror.Mask(awsErr)
			}
		} else {
			return microerror.Mask(err)
		}
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "ensured bastion ignition object deleted", "key", objectKey, "bucket", ignitionBucket)

	return nil
}

func (b *Bastion) ensureSecurityGroupDeleted(ctx context.Context) error {
	if b.bastionSecurityGroupID == nil {
		return nil
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "ensuring bastion security group deleted", "group_id", b.bastionSecurityGroupID)

	i := &ec2.DeleteSecurityGroupInput{
		GroupId: b.bastionSecurityGroupID,
	}
	_, err := b.awsClient.EC2.DeleteSecurityGroupWithContext(ctx, i)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidSecurityGroupID.NotFound" {
				b.logger.LogCtx(ctx, "level", "warning", "message", "bastion security group not found", "group_id", b.bastionSecurityGroupID)
				b.bastionSecurityGroupID = nil
			} else {
				return microerror.Mask(awsErr)
			}
		}
		return microerror.Mask(err)
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "ensured bastion security group deleted", "group_id", b.bastionSecurityGroupID)

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
