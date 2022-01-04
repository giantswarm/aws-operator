package template

const TemplateMainIAMPolicies = `
{{- define "iam_policies" -}}
  NodePoolRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: gs-cluster-{{ .IAMPolicies.Cluster.ID }}-role-{{ .IAMPolicies.NodePool.ID }}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: {{ .IAMPolicies.EC2ServiceDomain }}
          Action: "sts:AssumeRole"
  NodePoolRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: gs-cluster-{{ .IAMPolicies.Cluster.ID }}-policy-{{ .IAMPolicies.NodePool.ID }}
      Roles:
        - Ref: NodePoolRole
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

          # Following rules are required for EBS snapshots.
          - Effect: Allow
            Action:
            - ec2:CreateSnapshot
            Resource: "*"
          - Effect: Allow
            Action:
            - ec2:CreateTags
            Resource:
            - arn:aws:ec2:*:*:snapshot/*
            Condition:
              StringEquals:
                ec2:CreateAction:
                - CreateSnapshot
          - Effect: Allow
            Action:
            - ec2:DeleteTags
            Resource:
            - arn:aws:ec2:*:*:snapshot/*
          - Effect: Allow
            Action:
            - ec2:DeleteSnapshot
            Resource: "*"
            Condition:
              StringLike:
                ec2:ResourceTag/CSIVolumeSnapshotName: "*"
          - Effect: Allow
            Action:
            - ec2:DeleteSnapshot
            Resource: "*"
            Condition:
              StringLike:
                ec2:ResourceTag/ebs.csi.aws.com/cluster: 'true'
  NodePoolInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      InstanceProfileName: gs-cluster-{{ .IAMPolicies.Cluster.ID }}-profile-{{ .IAMPolicies.NodePool.ID }}
      Roles:
        - Ref: NodePoolRole
{{- end -}}
`
