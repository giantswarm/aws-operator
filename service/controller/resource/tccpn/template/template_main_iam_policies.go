package template

const TemplateMainIAMPolicies = `
{{- define "iam_policies" -}}
  ControlPlaneNodesRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: gs-cluster-{{ .IAMPolicies.ClusterID }}-role-tccpn
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: {{ .IAMPolicies.EC2ServiceDomain }}
          Action: "sts:AssumeRole"
  ControlPlaneNodesRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: gs-cluster-{{ .IAMPolicies.ClusterID }}-policy-tccpn
      Roles:
        - Ref: ControlPlaneNodesRole
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "ec2:*"
            Resource: "*"
          {{- if .IAMPolicies.KMSKeyARN }}
          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{ .IAMPolicies.KMSKeyARN }}"
          {{- end }}
          - Effect: "Allow"
            Action:
              - "s3:GetBucketLocation"
              - "s3:ListAllMyBuckets"
            Resource: "*"
          - Effect: "Allow"
            Action: "s3:ListBucket"
            Resource: "arn:{{ .IAMPolicies.RegionARN }}:s3:::{{ .IAMPolicies.S3Bucket }}"
          - Effect: "Allow"
            Action: "s3:GetObject"
            Resource: "arn:{{ .IAMPolicies.RegionARN }}:s3:::{{ .IAMPolicies.S3Bucket }}/*"
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
                autoscaling:ResourceTag/giantswarm.io/cluster: "{{ .IAMPolicies.ClusterID }}"
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
              - arn:{{ .IAMPolicies.RegionARN }}:ec2:*:*:network-interface/*
          #### Used for EFS
          - Effect: Allow
            Action:
            - elasticfilesystem:DescribeAccessPoints
            - elasticfilesystem:DescribeFileSystems
            - elasticfilesystem:DescribeMountTargets
            - ec2:DescribeAvailabilityZones
            Resource: "*"
          - Effect: Allow
            Action:
            - elasticfilesystem:CreateAccessPoint
            Resource: "*"
            Condition:
              StringLike:
                aws:RequestTag/efs.csi.aws.com/cluster: 'true'
          - Effect: Allow
            Action: elasticfilesystem:DeleteAccessPoint
            Resource: "*"
            Condition:
              StringEquals:
                aws:ResourceTag/efs.csi.aws.com/cluster: 'true'
  ControlPlaneNodesInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: gs-cluster-{{ .IAMPolicies.ClusterID }}-profile-tccpn
      Roles:
        - Ref: ControlPlaneNodesRole
  IAMManagerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ .IAMPolicies.ClusterID }}-IAMManager-Role
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            AWS: !GetAtt ControlPlaneNodesRole.Arn
          Action: "sts:AssumeRole"
  IAMManagerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ .IAMPolicies.ClusterID }}-IAMManager-Policy
      Roles:
        - Ref: "IAMManagerRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Action: "sts:AssumeRole"
          Resource: "*"
{{- if .IAMPolicies.Route53Enabled }}
  Route53ManagerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ .IAMPolicies.ClusterID }}-Route53Manager-Role
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
      PolicyName: {{ .IAMPolicies.ClusterID }}-Route53Manager-Policy
      Roles:
        - Ref: "Route53ManagerRole"
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Action: "route53:ChangeResourceRecordSets"
            Resource:
              - "arn:{{ .IAMPolicies.RegionARN }}:route53:::hostedzone/{{ .IAMPolicies.HostedZoneID }}"
              - "arn:{{ .IAMPolicies.RegionARN }}:route53:::hostedzone/{{ .IAMPolicies.InternalHostedZoneID }}"
          - Effect: "Allow"
            Action:
              - "route53:ListHostedZones"
              - "route53:ListResourceRecordSets"
            Resource: "*"
{{- end -}}
{{- end -}}
`
