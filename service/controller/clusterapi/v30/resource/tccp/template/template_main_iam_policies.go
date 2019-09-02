package template

const TemplateMainIAMPolicies = `
{{- define "iam_policies" -}}
{{- $v := .Guest.IAMPolicies -}}
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
  MasterInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: {{ $v.MasterProfileName }}
      Roles:
        - Ref: "MasterRole"
{{- end -}}
`
