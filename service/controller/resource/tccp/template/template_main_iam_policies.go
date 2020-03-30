package template

const TemplateMainIAMPolicies = `
{{- define "iam_policies" -}}
{{- $v := .IAMPolicies -}}
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
          {{- if $v.KMSKeyARN }}
          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{ $v.KMSKeyARN }}"
          {{- end }}
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
          - Effect: "Allow"
            Action:
              - "autoscaling:DescribeAutoScalingGroups"
              - "autoscaling:DescribeAutoScalingInstances"
              - "autoscaling:DescribeTags"
              - "autoscaling:DescribeLaunchConfigurations"
              - "ec2:DescribeLaunchTemplateVersions"
            Resource: "*"
          - Effect: "Allow"
            Action:
              - "autoscaling:SetDesiredCapacity"
              - "autoscaling:TerminateInstanceInAutoScalingGroup"
            Resource: "*"
            Condition:
              StringEquals:
                autoscaling:ResourceTag/giantswarm.io/cluster: "{{ $v.ClusterID }}"
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
          # Following rules are required to make the AWS CNI work. See also
          # https://github.com/aws/amazon-vpc-cni-k8s#setup.
          - Effect: Allow
            Action:
              - ec2:AssignPrivateIpAddresses
              - ec2:AttachNetworkInterface
              - ec2:CreateNetworkInterface
              - ec2:DeleteNetworkInterface
              - ec2:DescribeInstances
              - ec2:DescribeInstanceTypes
              - ec2:DescribeTags
              - ec2:DescribeNetworkInterfaces
              - ec2:DetachNetworkInterface
              - ec2:ModifyNetworkInterfaceAttribute
              - ec2:UnassignPrivateIpAddresses
            Resource: "*"
          - Effect: Allow
            Action:
              - ec2:CreateTags
            Resource:
              - arn:{{ $v.RegionARN }}:ec2:*:*:network-interface/*

  IAMManagerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ $v.ClusterID }}-IAMManager-Role
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            AWS: !GetAtt MasterRole.Arn
          Action: "sts:AssumeRole"
  IAMManagerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ $v.ClusterID }}-IAMManager-Policy
      Roles:
        - Ref: "IAMManagerRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Action: "sts:AssumeRole"
          Resource: "*"
{{- if $v.Route53Enabled}}
  Route53ManagerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ $v.ClusterID }}-Route53Manager-Role
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            AWS: !GetAtt IAMManagerRole.Arn
          Action: "sts:AssumeRole"
  Route53ManagerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ $v.ClusterID }}-Route53Manager-Policy
      Roles:
        - Ref: "Route53ManagerRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "route53:ChangeResourceRecordSets"
            Resource:
              - !Join [ "/", [ 'arn:aws:route53:::hostedzone', !Ref 'HostedZone' ] ]
              - !Join [ "/", [ 'arn:aws:route53:::hostedzone', !Ref 'InternalHostedZone' ] ]
          - Effect: "Allow"
            Action:
              - "route53:ListHostedZones"
              - "route53:ListResourceRecordSets"
            Resource: "*"
{{ end }}
  MasterInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: {{ $v.MasterProfileName }}
      Roles:
        - Ref: "MasterRole"
{{- end -}}
`
