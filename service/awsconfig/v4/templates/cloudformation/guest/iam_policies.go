package guest

const IAMPolicies = `
{{define "iam_policies"}}
  MasterRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{.MasterRoleName}}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: "ec2.amazonaws.com"
          Action: "sts:AssumeRole"
  MasterRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{.MasterPolicyName}}
      Roles:
        - Ref: "MasterRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "ec2:*"
            Resource: "*"

          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{.KMSKeyARN}}"

          - Effect: "Allow"
            Action:
              - "s3:GetBucketLocation"
              - "s3:ListAllMyBuckets"
            Resource: "*"

          - Effect: "Allow"
            Action: "s3:ListBucket"
            Resource: "arn:aws:s3:::{{.S3Bucket}}"

          - Effect: "Allow"
            Action: "s3:GetObject"
            Resource: "arn:aws:s3:::{{.S3Bucket}}/*"

          - Effect: "Allow"
            Action: "elasticloadbalancing:*"
            Resource: "*"
  MasterInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: {{.MasterProfileName}}
      Roles:
        - Ref: "MasterRole"

  WorkerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{.WorkerRoleName}}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: "ec2.amazonaws.com"
          Action: "sts:AssumeRole"
  WorkerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{.WorkerPolicyName}}
      Roles:
        - Ref: "WorkerRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "ec2:Describe*"
            Resource: "*"

          - Effect: "Allow"
            Action: "ec2:AttachVolume"
            Resource: "*"

          - Effect: "Allow"
            Action: "ec2:DetachVolume"
            Resource: "*"

          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{.KMSKeyARN}}"

          - Effect: "Allow"
            Action:
              - "s3:GetBucketLocation"
              - "s3:ListAllMyBuckets"
            Resource: "*"

          - Effect: "Allow"
            Action: "s3:ListBucket"
            Resource: "arn:aws:s3:::{{.S3Bucket}}"

          - Effect: "Allow"
            Action: "s3:GetObject"
            Resource: "arn:aws:s3:::{{.S3Bucket}}/*"

          - Effect: "Allow"
            Action:
              - "ecr:GetAuthorizationToken"
              - "ecr:BatchCheckLayerAvailability"
              - "ecr:GetDownloadUrlForLayer"
              - "ecr:GetRepositoryPolicy"
              - "ecr:DescribeRepositories"
              - "ecr:ListImages"
              - "ecr:BatchGetImage"
            Resource: "*"
  WorkerInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: {{.WorkerProfileName}}
      Roles:
        - Ref: "WorkerRole"
{{end}}
`
