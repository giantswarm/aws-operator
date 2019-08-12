// +build k8srequired

package setup

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
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
			UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(`
				{
				  "ignition": {
				    "config": {},
				    "timeouts": {},
				    "version": "2.1.0"
				  },
				  "networkd": {},
				  "passwd": {
				    "users": [
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "joe",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDE178xsxTfHERTXpzxbd8AsH4l1kQ+2y2+s1Ed0YQTfbNzCHMBKuCmabyv56QISc0Frp6oFNutmbRQpRlxNRzvWvcdapb2+wNQIOpc6/aQBPbiyCdU6Tjcw1p3p7z/O8M9wIPZ2e9zYyUjV0EzN/iZxrdDBduF1mrAjKzeG9E+McEUaD3LIJCxmljrt3248wusHvdwpLJGTM8K8ajdrlKNET9KEI3lWTaHBxr8v/cPixBJb+rxnMZuBRV/Hc3XN13OhW3wVftGMkgjrS0oVTcXE8YlrCYCNNlw+A1hVHZ3XBbV/g1Ww65lmL2AOHrOlnUd96bbocFcm6btqUuWr1clDfEZ/FvfAvWKe9pZb2rCxqOCnLzZmB6zUPj9dS8Cg7nnXZFfsIE0p71sO2i4cYd0l9uzQpmsxYPAy+BAdRpR9P2oM1CM/DbLjlO5IIb9qTB3O4R2zaG5WpVjAdvqo9XptXKa5uIi8ZoVHvhCdnqskwaXsMpEHavQVvdxPBal01smXxFv6lLqKMVkzJRBkXBEWXvxa12pv2kiFnaxMWK95jqLFHXpjZVrYS1Z77ld9+SXGr0KjAvd6SShPg1ggiDAd4suBDUbeyVQyhzr0CGJ4auiqHsO5IDSdaFFo7xeqnBzAT+jxsBfzKhn9In7HZNf1XnG+2fF41eqnobWwMbCaQ==  joe"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "giantswarm",
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "marian",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDMSBIK7+MCULpHyUhLvYOKgQEyhkBR1n7/5oyKhzAD5pvAjqBzBcQ64S4EYxL0/Y/JDa2bMzAEzaw0VL1e2kz/nAfd7QiW3ZgyU6uYGDwHHWDekAY+Q30giQoqP3QxFSDTjUVb1EC4kIO49uzAwItwM2ah6C/Jmz4/EWMP+2MKrwCe8DUTCYPI0RyXpyj0O0Uz+11VGVCIdMbxq3O62giC4WwNUFC+RDGS4plrsOo4whrLOlE7ZrYjSp1dU+GdNQmrKXJA8j9k8asIsChljrx6wF2aS7gMF5ltj8M3ufk1Cz4FN8/5luAE0qx14I8K0yej8Ann4dohrRm8sPz3aQOh marian"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "puja",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDFKCoKe/Z2cT7duLiiPdgUXKcsHx+3ESCa0t4hOZtfw6BHptJ0dpDTAqbkvpGRwErVpOO9tIQittWnzgX0RLqnDB98I7eZQ04AIZwoW9AX2aMEyLgbVMTeG9Xgy79pefjVV75T2lVjXtcOY2wJtf6ZU1KFq8/dHZb7vYzzobHBq9j4vIB5ZsNI94jJm7I6TLr24vga3+MEgQrsEQdRqZ2vxacSU1h+LSdfseGQew1XlxSTfTfglUcXE0WUlEFnak9z0JwQbblEmKQsinIwO4O0Sk6FQXObCFlss//gubk64/OM/87I/aKjrmbQTRMkxyqJ5jO4yIXOxeHpp5kNA9AKSmgHABhr1ViS6ocWO8mMekbLdxDWMdViTR6XxtFSPUCgTFAirsQi6/9qfV+6u2RLhKihuajy8akFi4BYqSGr17/crrkCYydBJRUIqNmQSdzGKodTJ9d556iFZ3rCM+Xe2mm4KsHkIQ3YphPMzb0yAWEtZl1ncdqSXHz74M9b1KHUzyJgQhv5KOzhURxXR/UVBy7NNPae8XSEFId/O2uHgc/mWV5Xr5ZwxbwXsmlyto53+EmgynnPcL96RgVyiAmHL/vtvOGAzVSOPtNsnU7QG/YDfA2WrxLmuGEA0WzC63iXZCqSFbPK0adelJo9vctCB2gozrVpjqXWpskg8SZMqw== puja"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "teemow",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCvDv/5/bOgnuNXbxB7n8n8reDpRdAPeZvuCE3pJPVYJNfL0hoNzGhw4MBLRivSd9ViMM9X1CE7iVK6RFO46yt4GK5JWYi1UM88q6I+4bcnswUAWsIFX5a/U0LK0wIR3akDDyU9WEA7CXoJ7IxF7dtIYC9OIqrD6gXc4P/UI0jZQI5iZY3qNjlKVAwsNz8pD/BE2sPpNVHumzgcLJEveoc3WMCmfBAAWQAMfRlhlJ5LjM7Py/5k1/s5Myn4L/yoAvxMWev4k2ZYpumt752/r927K7AIrK/OYTfqLPKzZYSLWAj4g7L3/65sKpFm6g+HFgDmlScgf9P4bAoLn6+mWbYL teemow"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "tobias",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCS/thnoYaUYmuDxwQF9ES3Jsq6KltO+QU/8PVo1tUr5vlEfEY1Q9JYHiPKJW+U0cMH3a/Jv2IDTaH629djoNdtTottaYGDINBoVIlAdR+vwm3JzVUB02mb+QXTzhzLc58fdwhHN0PS82/BcSSFpQzI7PedRGMtzS6Qxcx4YfrzC16vsdF8wMw+tVbtI2fDLwfd9NcpCDr582NrX/qOR22zeck3VVgppnuC5mGAC+XvHCRbp+4pZJ0W4fpEIGwG1cPbktvdA0wYcn7GJo7fU11066PMGMXplV+DEnQTpBUbP+NFXRuY7RzTeuTGSZHXsWO11cmpLPVVB7LdAaQuSPi1 tobias"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "xh3b4sd",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQClCCgsKl7+mQwD+giN6OEruV1ur/prpWXfyGHJyGGQkROZA3IcrpmRPWmKKXpCaW+G8lcb9DXD/K7/rNAh+4hpsfvCUs8u0mJ6u4El/8dcRTQaZUdLX8q3AZZ38gmk+yZz241x7LGd05D4H+aq9sVdtbcAepINUJyZ7p3yXTfCYwHC7QMYiuRFKMaUHY50shFhSYdD9TCEFtH2ybPi1/WOCX6gf90f6O0Ivo7tzwtYGV8ToIa2nO+CqwlIRiGqEy4/g9h1gCPDvgcLZmok74V6mH12whNdMDyJyuT8S1dLwNiKoYkvMbcUkpE0O/0LBCg+SsHVHmgnsNx9t0hUg8iR xh3b4sd"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "calvix",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC9IyAZvlEL7lrxDghpqWjs/z/q4E0OtEbmKW9oD0zhYfyHIaX33YYoj3iC7oEd6OEvY4+L4awjRZ2FrXerN/tTg9t1zrW7f7Tah/SnS9XYY9zyo4uzuq1Pa6spOkjpcjtXbQwdQSATD0eeLraBWWVBDIg1COAMsAhveP04UaXAKGSQst6df007dIS5pmcATASNNBc9zzBmJgFwPDLwVviYqoqcYTASka4fSQhQ+fSj9zO1pgrCvvsmA/QeHz2Cn5uFzjh8ftqkM10sjiYibknsBuvVKZ2KpeTY6XoTOT0d9YWoJpfqAEE00+RmYLqDTQGWm5pRuZSc9vbnnH2MiEKf calvix"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "rossf7",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCY0Rn2VqhtOFy7LY6nu53c9bP5Fy0KQ+P5/MA/s4aG8+veIqTyhhpHLwPF0hKbi1mq+HaFvm1nLbovZcTTG2Z+BoP/Y5kV5EOjnq415EtT3rH0YdM0h69Qxuc0KqUvU/F43XOhpEH0o8L+ZK+Vq4UrRPIDRjftc8N5h6MJszAow/kiC7d30nYsPio6FuWHH5jZaAKjucQbBsDU5r160mtVk0HCexutm3s3fHTADojZjFA3t8FJy7vO+Og+VDVzV9ai9E32mgytNL0wVE1dUGqPwoM9MrzxNC2TZedS74zqBoK9TL3y1sfVzD5RpdX4Z5FInhtTz1z4nnYzsPiYQKMx rossf7"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "oponder",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAxEDA/Nr7++9SoFOhTeNJMAubkzJmGZWHtXul0kM+FJR0TM/1V3b4XfoRAwU9gaG88P+4venSwcbLVfvvacnlGQ8hTgw2Jlpz6Z9+iIVjru2+nYgJeELFff5bdPLYWE4Ft/VYpwGibD1DbVGldsO3I7EdaEfd5FeOF0Fk1xPK50UGvTq9CkU+wEcTc9eDzFWpLoz/69KWG/F7XEZhWqswUjHaN1UJtuBlVmoe/0OrlyIYBl2CeUzpmJHNDDv7u8gKOCOFwmzMdDieGHzIITCovIoGLhIJS1dGdT3FPvTPG/s8VOZ/QGS9rfsc5x08J0v0JqwN8GZT9FxFJECeXTN6bw== oponder"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "kopiczko",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDzY+7kp4XTNfinVLbDo28E0yaJUMvEabzdsheGHG3+gubakJgITZw+3m4vy45WRoF8QDOjpZ12n9ov9bpz5X8kRmHSthvWYtNRYJWCJc0d3+td1/Ki9CaHesNhKdeVYcw9g5x55h6o4EPx+g6wIsBhxqjcdZ5O37M+KWXlBfLoP4WKBjORhD4kpU+suA+rMIRF/njLs8zswHL8or3Ynp+voZM1PVCfxENp9ktNeA7W2KyUZpgDtoWxN2cnj0BOs/t2w+XZhqgsPo/9zXO6C0XIvPv69MAOHYMsomKldQgpy+MlODvu/sbP5ruB/4vGiqCg0+E2KS1ZK85KytQtfai3IcmglNtHdRKyYmf2WkY/GMGB6sgQKZWJE4XeN7+mZvRrOih2yo+/GJCkI5U0eWIUInyh/lBMIxVbckbpdZYxMd2SaERkTFNDhqSgncrPu5gW7ZFejgie9gVlkYBLI/D1XXM1YzvpSlA80D9kxLhiqdaAXP7CysPgS7EN56zM0SzHR0vxrr4dhB9XuBxlMeTtC0dvaMPkiIJj43MLDNzE4wtXEpzFQnmzRoOrPAACRhr0OcOTbsH7X+QR9JUB1ygTRSN/kHVFyL+EPRRsfuzbxw57jkzIE7ihGtcINtyv4dIFfqE7c+uDDZV/yUi2C5PI6FLHW33DXwqbzR8yZuOu2Q== pawel"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "marcelmue",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC9pOt/FFXonkNDGamoGMVg6wQJYS7m9r/OO2wWoEPNQS/M4nGL/szKlZfD4Z6tKS3WyeLY5XU7QjOFb3Gt3QQAaVOSDgTkfH1i+usWSNzFlhjARQkIUs0j9m30o7sXznZNOy4r73bnfUjTwYJifGUWBPq/jGokLNCxBRPCaJFi8Od5De3DyuDd93SAoXkTJaDPr7J90tzkVLI6ek5va1GSTeVdHbAifds8r8Shm0wgdmcVKiOYt7/oyzavl0x5XPzMAVXeUI0jIopsvqjiy/fS+Cq7i1TMBQ+rkycWLc8X8CM8U84OQ6eb8LgQw0A4xqVtZHOl2FlHHtWNjLhnwO5MHWjdWxCUSEshK+tS5Wm64ph37nfObupPpoMcRRTmKB2SdsigU0T/aJ9zoJorsIBKDY6lqmXoYk07XEyu5tmuG9cuiP900yLRIoCZeQ14WP3+3KDWLfjic3W8wXr0xlLOTaPNtNpX+v6X4n8R4HuJ0zQw1znEXhBlhTEZTRe1qmdPQTNwQa6XwOlEJVYOIHEOobTQVu+ReIrj+XT3b1VR531pG8M3o9Rhq29aFpAsYiDh12aGk45b+61hddF3WybpMOVQqfvYnf1lVwt/0PujJuIC7e6WuHAlRS2Sshb19ROG8w2mW1sGbp2zN7Y+MUAI5LVOHrBtRqEoZkhz+JXMzQ== marcelmue"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "vol",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDFGDg/p4JWewXAs8kExJnCaNXEN1v2LZf0YWWiblHFp1+i2bp8qSmAJT3i6Yw0kHY2/6MotBCKAsFtlqxuhKaFs3jDcmdOugmWz4Qj7oerQ/ypJE/wZ9PY79gbK75aEKyOdVf7dUT6Ah+oSfETgpY/3a9pVZ/dSF3WBFIBw5k4YarFzcELQE4Bo4dcsLHsNrkI9Bk6gkGbTY+1TtfJmOu0bEXxXHdEq+JfW0MFssjh3I5n0DT09qDnztAvRAjjqjlyNKNt8reErV0LlvsDM5c+426Bz9JgM5vP3sD5ai8lpuH0iCBHoo9678XTKKTYbbz0s7kgXUb0vGS+GbOcaKBKmZ8a0xDpsft9+/LbmnuUic8b4c4/cRw5wSV1IYqyDqARp/d9PaJlYa22ISGnDbYmXUTsef0PhUenK9gtYrGsVhQmkqeLYiIYqwsl7+uouFMpQDmdZjY/B4fKcRA3oRGCFuwzT1vrtJL41dw9WyzM+3xnHTMFZdko9TlgDiEeu6gdpsTGJf4VALUWgXeyW/egte2im86kjMxzQuCw/aOmiYMqwZH2YfI0dS9jLuZbxePKTUounct66SrNXBrbu2d0BiPj6bl1dG6oZhwtArRnbiG5+cTakDvLhFgahTQFAT1De7o3Nr+BfjNQkVlQNKaIPUOdypiDNJE/6q/GOHVRQw== vol"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "jgsqware",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDg0b+m551vLRqsnnDrB32PrQnO41Fv62dbYYCGqjcd/if8iOuyXvxNRnbOxnsFFPVSiqY2LB8pXofEAm5MG2qupGyBvivtnq1v+6Ld9nMYTaKdz5WKhI9ypQ/jV4G1DNYrayGno13eRmGemCEnIdZeRrVxp5EfkVX0ZyJ88998Urjv6OtSLV+GSNSiIbNYyvGjLoR0dt5LCVbwbbQd1H5wXYsSoeIkKiqfS7AtMn9wDCIyM1W15yC/4UaCMEGkVfjLZB+4Y8BBfLH1vI1h43zl4EUkaq1PASDvpX0AWlclfemkK2bGEOS/UzVJsZtM73jEoSZvq/aCLe3v/0zI/5Xl julian"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "fernando",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDdprdoupVo8EH7yL3BDwj1odnJfVs6ab0BZehepcVGL7Yno5YhTM/cof/KvGWZyzER1m0hn8adm9j6HqjhskzhoUSq37b0q4v38t1oZZ7YPjF5b90bvQKnUv/5U6xqOFnE9LfotLFT8cidLLwPSDKOlPacEmSheDKTRFZwAfCJuv4eobZ3Hs0qCfzx7IjKqxYWfEJSvp9x2EUHBJnBSZ3n1ncbpyI6wyyXgjoo9v7bXhWu+HRB+LyXCPvDME7ZWihNdlX4ZesiEoyOtoVKZsBx4/L1ckyhPN2NNDNElG41/w8znrbnKBKPiBM5AGnc7XZw2TP1ivWhJ2EVBjvtPT73GMgR6AdIV4iInNdg89mLKNtHwqZ8Gyj7/xS7ufdMFST1EGnBnk/mGvK1eITWVv6Y9s5Zte84LRhEK7W2jQtND70lT3FMBKLVO/W0q+suWZIqAYWzfpASnvaXlLWIrRApzslzgbyP14Zdp1cTzjwXOoQ+FhHRlqhX7G8uGR7JvNaS732kD8xASDzZulDrkmtxDVYqKFTxRQRwolmCeDzokS2OdnRncS7m6AxwMbwI0QuQwQZp3+G0qIxKXCad+XPCQt2oaYJJRZD+dqPJOKSofbS3eZayP0nPVNqXBP55S4H2PGGSATIpSRQdfy0bebV54x7N+P3KBjxzvIpzX0nn4Q== fernando@giantswarm.io"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "tuommaki",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCu2V1p0cZ2N4ucug9LQMB+YMg9AQ5+aZQTdDTZ7oBuEcBuGtdnSSbcxj1lHoMYvhz6ugFVolkusRnZSakZY/XPVlwIHC56TWWrJ0hJ4sQEzCqVSHx0ZBHaMZepxCz7KSh/4KjtZFyaBC9SFwUo7kGgBYoFdClhxZsmfMsk0RneY8FjWme/cwXSaEGdaaTyOA52UOCg6Ax3nnE/gAJBsL8HgI17bFjj8og6TdPoP+33wujGHFORy8HF/m6p1I2Nm9Mp+gkG6PzdkWbF7UFci5uYHXy5IEu6uGzEPQiB5BjgfVIvZyH3VfKxmG1T2yyp4/qDQOmkjlIahpPyI00Y3SWAab7MdQXJ2hTgWFo/NP+AEdd45+PrSvTMy2k5bVl9GMntP+z+9oAhwH8OStSCJ0GBGlVG89fd0vFV1XVmLPwS8XhuhAoU1KRt6/Hc8cs7uSUiKOTY8Xn6VNUozxK137QpHBb81jU7OCcmopF9dlqoV6m18iZK1NjP4+FFxUyi5O4HI6aFrZXf7Cw5G9C8EXML3qLIMxd2pIJsu8QTw/5kC7sBtmFY/5RqW0TZ5hWuyGSuFcRan5E08Qct5rGAQ6QjJ9rZqQUPeJFcN6gEvGUam0XdeziZD6lPFUDkte9y653lIrPqBoSbsJuk/FJU/+RTSYEl+VCmaac3ru6jYV6M8w== tuomas"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "theo",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDo5PU8w8RDWYDd8SSIKOiYJCeN398PAEJApFGpPWprewBmiDGGMHHDIDV//QR+o5MhR2hBmJ+Pw1K/Edr4u0cfGlIIb6lSVdha+jDEp0l1PtqyQzubzTH3y/RDzxCakAa419N1G3DwzJHkWBxbpyqx7i/DOYcxgQP4vCGvYgkvuOkQYCNk2hfuXAl/Aucv3JXlsuNktyXQ6XKf+Twa2Bg8jIAaYUNGqgKgzMcsCElE55bxVuYeXl441CzD2fdHmXyGwo6nefN7PZ790SxQzkGM8wBpESgc7U4IPUY/dnsn4yQBYw2meontHWGLmZjrvEYxoS7Uv4o8BX8cScgVZUhRojHvNWBBcwOx+hhuPbqoqdc8IFXQLHTTa/szvY9gwlipBejj6nJrRpl3Kxw+EX4QP/loDxkykWQUoByU479Z6/gOtgAkPOe8xZblny6r3uCZyUlaYR9ht2aOEbH8bLuYBaDTPvunMIH/RSgbNxys/Dss3ZC7MJgXtoaSpb/AGdqv1Uj4AdNJJm8544AFhmR/Tky0rms3NpaSEiwO+E5ZIJiBevqPbGWkodbfKM6uydS+wrqHOR+zTNLuwhTHVnNRZ//ePBMptnzR1qbuMmaEgmqMM1HjFflUUVdDLFK6TxcdU3YPJnnWlwGk2bCjlamGjHx0hvoJVSY+Oap6o23cJQ== theo@giantswarm.io"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "jannis",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCrHyOvY5nH1DCybBDcA0BdZ916zrtaeP1fNBaqbxsfCg3m8SjMZVb62W5PHI5177o4Ci4JoMUj/+HpVY07Ao5UXCJ3J+LXPNaAsyzfijHBpVas7YgWG9pv3OtElFnB3wYPz1VpJH2BCy6p1SomvxqyJ6ihLeLIzWaXHoGu5d73aeKXKEelrpvl5Y2wFPSQBqIbl7prRB+q4ChG+D6ZR/9Zd8Tzth7WUrimJbmDUWBUKmhfgEYTjKVmIuUrHA2yRBv6dRE1yP9A3W5A8fOHMlcyEsKw+ZK4i7LZn+GRVY144K67w38B85KlPBmhVEXrbvFOLXZxzZqg9glb1TnVZ7eGzJBH3B72GB7y/igEytaGcHw/qt53+B/zYPTJFLC9fpyaLWC4JswzyOlNlybMZJY6EFxlvb+c0ALt935fhq+3+20SUIikZsTbQGMxnaTI7V48PMTxngNRykdBMYleB9iNQ6Weh2Ox8ZJFaWjKTGgtVujMtJEHI0rDpB8GnVgu0/No36z32E6juBApDCRLuoJDnNd47z6yM4Zq+bKL/xUIp5gTNKdiD1hVVrBv/HoECtHZjRrXklX0+wRPQV2AuTdl7kK1IfRtLet9osQgpEDVVD91D3xgPX4W2n2F180F6wgQLMzQuzcHtlepaDv459D2Mac1krkKbnNC131ElPfIMw== j.schaefer@estwx.de"
				        ],
				        "shell": "/bin/bash"
				      },
				      {
				        "groups": [
				          "sudo",
				          "docker"
				        ],
				        "name": "thomas",
				        "sshAuthorizedKeys": [
				          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC+0XcR/g2ebbzSm3iO+DgceVRm9ZpnKe6WDc+OjTSzupa4Y/nhROEQZtNEllTkkgd7lraACdlu6PPf3vuqoaA/XZ2NNYtSOJ19JvZclLKaT7QwGlD4fMgHxESeo43SSH8Hpn+aZYgLYPjBA7lSshEpv0DSpeKhoL7sD9+SonGklRjLrhpr9RlGBlds2UiuNEVrp0TpBEqWacxPrlZ9UYfuU6lGuMWC/735ShvGMusghB9q/HLd8LdwkhY/yA31Hspc03rQbCUNC27vaUEEpGV27Z02b0e9xloo5KCHNDLBNP7zCDPCwAmwzvOkrxGAWR2fabBSvCjMBUsaIzI3PkmwUm8+LJnRnbbWTtOfcsJA+L7pptZ51d8lYenKfaszgXSKE873LxEXOMfZv4Jyzn6mBayCMFYpM88gle3nl6vjwVaE4xm5mAphZsxr7+PXhHpHtFx3dvsdLRN6vnC/woV3eY+2yj1zwukvp3P+h+su1moR2DfxBPFEhZTp2jtUtaSVwm787rtDxMGlxeZ8spLv1x2tquTDfozlM88nC7Q5nqLQcp+Cow9XqFPn8shgtboWeIc9U4zJe1D0MiJSdsvn2fxn6Q7UB65yY2xeKzkb1O3c/4V76X5q/CYJkX7/E1WdHWNlm9CQd7fq6EGBh2v4TfDdkZNa0iRJ1EPs33/wDQ== thomas@fussell.io"
				        ],
				        "shell": "/bin/bash"
				      }
				    ]
				  },
				  "storage": {},
				  "systemd": {}
				}
			`))),
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
