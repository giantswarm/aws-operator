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

type bastionManager struct {
	enabled bool
	instanceID *string
	bastionSecurityGroupID *string
	workerSecurityGroupID *string
	vpcID *string
	subnetID *string
	ignitionBucket *string
	ignitionURL *string
	ignitionHash *string
	awsClient *e2eclientsaws.Client
	logger micrologger.Logger
	clusterID string
}

func newBastionManager(clusterID string, awsClient *e2eclientsaws.Client, logger micrologger.Logger) (*bastionManager, error) {
	return &bastionManager{
		logger: logger,
		awsClient: awsClient,
		clusterID: clusterID,
	}, nil
}

func (b *bastionManager) bastionExists(ctx context.Context) (bool, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			b.instanceID,
		},
	}
	response, err := b.awsClient.EC2.DescribeInstancesWithContext(ctx, input)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if len(response.Reservations) == 0 ||
		len(response.Reservations[0].Instances) == 0 {
		return false, nil
	}

	return true, nil
}

func (b *bastionManager) ensureCreated(ctx context.Context) error {
	err := b.ensureSecurityGroupsCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureIgnitionCreated(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if b.instanceID != nil {
		exists, err := b.bastionExists(ctx)
	}

	if b.instanceID == nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion instance")

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
			ImageId:      aws.String("ami-015e6cb33a709348e"),
			InstanceType: aws.String("t2.micro"),
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

		_, err = b.awsClient.EC2.RunInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "created bastion instance")
	}

	return nil
}

func (b *bastionManager) ensureDeleted(ctx context.Context) error {
	err := b.ensureSecurityGroupDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	err = b.ensureIgnitionDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if b.instanceID != nil {
		b.logger.LogCtx(ctx, "level", "debug", "message", "terminating bastion instance")

		i := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				b.instanceID,
			},
		}

		_, err = b.awsClient.EC2.TerminateInstances(i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "terminated bastion instance")
	}

	return nil
}

func (b *bastionManager) ensureSecurityGroupsCreated(ctx context.Context) error {
	var err error

	{
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

		o, err := b.awsClient.EC2.DescribeSubnets(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.Subnets) != 1 {
			return microerror.Maskf(executionFailedError, "expected one subnet, got %d", len(o.Subnets))
		}

		b.subnetID = o.Subnets[0].SubnetId
		b.vpcID = o.Subnets[0].VpcId

		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found public subnet %#q and vpc %#q", b.subnetID, b.vpcID))
	}

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

		o, err := b.awsClient.EC2.DescribeSecurityGroups(i)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(o.SecurityGroups) != 1 {
			return microerror.Maskf(executionFailedError, "expected one security group, got %d", len(o.SecurityGroups))
		}

		b.workerSecurityGroupID = o.SecurityGroups[0].GroupId

		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found worker security group %#q", b.workerSecurityGroupID))
	}

	// We need to create a separate security group in order to allow SSH access to
	// the bastion instance. The AWS API does not allow tagging the security group
	// when creating it. That is why we need to separately create tags below, so
	// we are able to find it later on when we want to clean it up.
	{
		b.logger.LogCtx(ctx, "level", "debug", "message", "creating bastion security group")

		i := &ec2.CreateSecurityGroupInput{
			Description: aws.String("Allow SSH access from everywhere to port 22."),
			GroupName:   aws.String(b.clusterID + "-bastion"),
			VpcId:       b.vpcID,
		}

		o, err := b.awsClient.EC2.CreateSecurityGroup(i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.bastionSecurityGroupID = o.GroupId

		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created bastion security group %#q", b.bastionSecurityGroupID))
	}

	// The AWS API does not allow tagging the security group when creating it.
	// That is why we need to separately create tags below, so we are able to find
	// it later on when we want to clean it up.
	{
		b.logger.LogCtx(ctx, "level", "debug", "message", "tagging bastion security group")

		i := &ec2.CreateTagsInput{
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

		_, err = b.awsClient.EC2.CreateTags(i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "tagged bastion security group")
	}

	{
		b.logger.LogCtx(ctx, "level", "debug", "message", "updating bastion security group to allow ssh access")

		i := &ec2.AuthorizeSecurityGroupIngressInput{
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

		_, err = b.awsClient.EC2.AuthorizeSecurityGroupIngress(i)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", "updated bastion security group to allow ssh access")
	}

	return nil
}

func (b *bastionManager) ensureIgnitionCreated(ctx context.Context) error {
	var err error

	if b.ignitionBucket != nil {
		headBucketInput := &s3.HeadBucketInput{
			Bucket: b.ignitionBucket,
		}
		_, err := b.awsClient.S3.HeadBucket(headBucketInput)
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
		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 bucket %#q", ignitionBucket))

		createBucketInput := &s3.CreateBucketInput{
			Bucket: aws.String(ignitionBucket),
		}

		_, err = b.awsClient.S3.CreateBucket(createBucketInput)
		if err != nil {
			return microerror.Mask(err)
		}

		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 bucket %#q", ignitionBucket))
		b.ignitionBucket = &ignitionBucket
	}

	if b.ignitionURL == nil {
		ignitionURL := fmt.Sprintf("s3://%s/%s", b.ignitionBucket, bastionIgnitionKey)
		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating S3 object %#q", ignitionURL))

		userData, err := generateUserData(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		putObjectInput := &s3.PutObjectInput{
			Key:           aws.String(bastionIgnitionKey),
			Body:          bytes.NewReader(userData),
			Bucket:        b.ignitionBucket,
			ContentLength: aws.Int64(int64(len(userData))),
		}

		_, err = b.awsClient.S3.PutObject(putObjectInput)
		if err != nil {
			return microerror.Mask(err)
		}

		sum := sha512.Sum512(userData)
		hexSum := hex.EncodeToString(sum[:])
		ignitionHash := fmt.Sprintf("sha512-%s", hexSum)

		b.ignitionURL = &ignitionURL
		b.ignitionHash = &ignitionHash

		b.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created S3 object %#q", ignitionURL))
	}

	return nil
}

func (b *bastionManager) ensureSecurityGroupDeleted(ctx context.Context) error {
	if b.bastionSecurityGroupID == nil {
		return nil
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "deleting bastion security group")

	i := &ec2.DeleteSecurityGroupInput{
		GroupId:   b.bastionSecurityGroupID,
	}
	_, err := b.awsClient.EC2.DeleteSecurityGroup(i)
	if err != nil {
		return microerror.Mask(err)
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "deleted bastion security group")

	return nil
}

func (b *bastionManager) ensureIgnitionDeleted(ctx context.Context) error {
	if b.ignitionBucket == nil {
		return nil
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "deleting ignition bucket")

	i := &s3.DeleteBucketInput{
		Bucket: b.ignitionBucket,
	}
	_, err := b.awsClient.S3.DeleteBucket(i)
	if err != nil {
		return microerror.Mask(err)
	}

	b.logger.LogCtx(ctx, "level", "debug", "message", "deleted ignition bucket")

	return nil
}

func generateUserData(ctx context.Context) ([]byte, error) {
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
