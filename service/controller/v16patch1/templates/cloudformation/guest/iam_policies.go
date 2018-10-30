package guest

const IAMPolicies = `{{define "iam_policies"}}
{{- $v := .Guest.IAMPolicies }}
  MasterRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ $v.MasterRoleName }}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: {{ $v.EC2ServiceDomain }}
          Action: "sts:AssumeRole"
  MasterRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ $v.MasterPolicyName }}
      Roles:
        - Ref: "MasterRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "ec2:*"
            Resource: "*"
{{ if $v.KMSKeyARN }}
          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{ $v.KMSKeyARN }}"
{{ end }}
          - Effect: "Allow"
            Action:
              - "s3:GetBucketLocation"
              - "s3:ListAllMyBuckets"
            Resource: "*"

          - Effect: "Allow"
            Action: "s3:ListBucket"
            Resource: "arn:{{ $v.RegionARN }}:s3:::{{ $v.S3Bucket }}"

          - Effect: "Allow"
            Action: "s3:GetObject"
            Resource: "arn:{{ $v.RegionARN }}:s3:::{{ $v.S3Bucket }}/*"

          - Effect: "Allow"
            Action: "elasticloadbalancing:*"
            Resource: "*"
  MasterInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: {{ $v.MasterProfileName }}
      Roles:
        - Ref: "MasterRole"

  WorkerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ $v.WorkerRoleName }}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: {{ $v.EC2ServiceDomain }}
          Action: "sts:AssumeRole"
  WorkerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ $v.WorkerPolicyName }}
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
{{ if $v.KMSKeyARN }}
          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{ $v.KMSKeyARN }}"
{{ end }}
          - Effect: "Allow"
            Action:
              - "s3:GetBucketLocation"
              - "s3:ListAllMyBuckets"
            Resource: "*"

          - Effect: "Allow"
            Action: "s3:ListBucket"
            Resource: "arn:{{ $v.RegionARN }}:s3:::{{ $v.S3Bucket }}"

          - Effect: "Allow"
            Action: "s3:GetObject"
            Resource: "arn:{{ $v.RegionARN }}:s3:::{{ $v.S3Bucket }}/*"

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
      InstanceProfileName: {{ $v.WorkerProfileName }}
      Roles:
        - Ref: "WorkerRole"
{{ end }}`
