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
          - Effect: "Allow"
            Action:
              - "kms:Encrypt"
              - "kms:Decrypt"
              - "kms:ReEncrypt*"
              - "kms:GenerateDataKey*"
              - "kms:DescribeKey"
            Resource: "*"
          - Effect: "Allow"
            Action:
              - "kms:CreateGrant"
              - "kms:ListGrants"
              - "kms:RevokeGrant"
            Resource: "*"
            Condition:
              Bool:
                kms:GrantIsForAWSResource: "true"
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
              - "autoscaling:SetInstanceHealth"
              - "ec2:DescribeLaunchTemplateVersions"
              - "ec2:DescribeInstanceTypes"
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
  ALBControllerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ .IAMPolicies.ClusterID }}-ALBController-Role
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Principal:
              AWS: !GetAtt IAMManagerRole.Arn
            Action: "sts:AssumeRole"
          {{- if or (ne .IAMPolicies.Region "cn-north-1") (ne .IAMPolicies.Region "cn-northwest-1") }}
          - Effect: "Allow"
            Principal:
              Federated: "arn:{{ .IAMPolicies.RegionARN }}:iam::{{ .IAMPolicies.AccountID }}:oidc-provider/s3.{{ .IAMPolicies.Region }}.amazonaws.com.cn/{{ .IAMPolicies.AccountID }}-g8s-{{ IAMPolicies.ClusterID }}-oidc-pod-identity"
            Action: "sts:AssumeRoleWithWebIdentity"
            Condition:
              StringLike:
                "s3.{{ .IAMPolicies.Region }}.amazonaws.com.cn/{{ .IAMPolicies.AccountID }}-g8s-{{ IAMPolicies.ClusterID }}-oidc-pod-identity:sub": "system:serviceaccount:*:aws-load-balancer-controller*"
          {{- end }}
          {{- if ne .IAMPolicies.CloudfrontDomain "" }}
          - Effect: "Allow"
            Principal:
              Federated: "arn:{{ .IAMPolicies.RegionARN }}:iam::{{ .IAMPolicies.AccountID }}:oidc-provider/{{ .IAMPolicies.CloudfrontDomain }}"
            Action: "sts:AssumeRoleWithWebIdentity"
            Condition:
              StringLike:
                "{{ .IAMPolicies.CloudfrontDomain }}:sub": "system:serviceaccount:*:aws-load-balancer-controller*"
          {{- end }}
          {{- if ne .IAMPolicies.CloudfrontAliasDomain "" }}
          - Effect: "Allow"
            Principal:
              Federated: "arn:{{ .IAMPolicies.RegionARN }}:iam::{{ .IAMPolicies.AccountID }}:oidc-provider/{{ .IAMPolicies.CloudfrontAliasDomain }}"
            Action: "sts:AssumeRoleWithWebIdentity"
            Condition:
              StringLike:
                "{{ .IAMPolicies.CloudfrontAliasDomain }}:sub": "system:serviceaccount:*:aws-load-balancer-controller*"
          {{- end }}
  ALBControllerRolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ .IAMPolicies.ClusterID }}-ALBController-Policy
      Roles:
        - Ref: "ALBControllerRole"
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          {{- if or (ne .IAMPolicies.Region "cn-north-1") (ne .IAMPolicies.Region "cn-northwest-1") }}
          - Effect: Allow
            Action:
              - 'iam:CreateServiceLinkedRole'
            Resource: '*'
            Condition:
              StringEquals:
                'iam:AWSServiceName': elasticloadbalancing.amazonaws.com
          {{- else }}
          - Effect: Allow
            Action:
              - 'iam:CreateServiceLinkedRole'
            Resource: '*'
            Condition:
              StringEquals:
                'iam:AWSServiceName': elasticloadbalancing.{{ .IAMPolicies.Region }}.amazonaws.com.cn
          {{- end }}
          - Effect: Allow
            Action:
              - 'ec2:DescribeAccountAttributes'
              - 'ec2:DescribeAddresses'
              - 'ec2:DescribeAvailabilityZones'
              - 'ec2:DescribeInternetGateways'
              - 'ec2:DescribeVpcs'
              - 'ec2:DescribeVpcPeeringConnections'
              - 'ec2:DescribeSubnets'
              - 'ec2:DescribeSecurityGroups'
              - 'ec2:DescribeInstances'
              - 'ec2:DescribeNetworkInterfaces'
              - 'ec2:DescribeTags'
              - 'ec2:GetCoipPoolUsage'
              - 'ec2:DescribeCoipPools'
              - 'elasticloadbalancing:DescribeLoadBalancers'
              - 'elasticloadbalancing:DescribeLoadBalancerAttributes'
              - 'elasticloadbalancing:DescribeListeners'
              - 'elasticloadbalancing:DescribeListenerCertificates'
              - 'elasticloadbalancing:DescribeSSLPolicies'
              - 'elasticloadbalancing:DescribeRules'
              - 'elasticloadbalancing:DescribeTargetGroups'
              - 'elasticloadbalancing:DescribeTargetGroupAttributes'
              - 'elasticloadbalancing:DescribeTargetHealth'
              - 'elasticloadbalancing:DescribeTags'
            Resource: '*'
          - Effect: Allow
            Action:
              - 'cognito-idp:DescribeUserPoolClient'
              - 'acm:ListCertificates'
              - 'acm:DescribeCertificate'
              - 'iam:ListServerCertificates'
              - 'iam:GetServerCertificate'
              - 'waf-regional:GetWebACL'
              - 'waf-regional:GetWebACLForResource'
              - 'waf-regional:AssociateWebACL'
              - 'waf-regional:DisassociateWebACL'
              - 'wafv2:GetWebACL'
              - 'wafv2:GetWebACLForResource'
              - 'wafv2:AssociateWebACL'
              - 'wafv2:DisassociateWebACL'
              - 'shield:GetSubscriptionState'
              - 'shield:DescribeProtection'
              - 'shield:CreateProtection'
              - 'shield:DeleteProtection'
            Resource: '*'
          - Effect: Allow
            Action:
              - 'ec2:AuthorizeSecurityGroupIngress'
              - 'ec2:RevokeSecurityGroupIngress'
            Resource: '*'
          - Effect: Allow
            Action:
              - 'ec2:CreateSecurityGroup'
            Resource: '*'
          - Effect: Allow
            Action:
              - 'ec2:CreateTags'
            Resource: 'arn:{{ .IAMPolicies.RegionARN }}:ec2:*:*:security-group/*'
            Condition:
              StringEquals:
                'ec2:CreateAction': CreateSecurityGroup
              'Null':
                'aws:RequestTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'ec2:CreateTags'
              - 'ec2:DeleteTags'
            Resource: 'arn:{{ .IAMPolicies.RegionARN }}:ec2:*:*:security-group/*'
            Condition:
              'Null':
                'aws:RequestTag/elbv2.k8s.aws/cluster': 'true'
                'aws:ResourceTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'ec2:AuthorizeSecurityGroupIngress'
              - 'ec2:RevokeSecurityGroupIngress'
              - 'ec2:DeleteSecurityGroup'
            Resource: '*'
            Condition:
              'Null':
                'aws:ResourceTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:CreateLoadBalancer'
              - 'elasticloadbalancing:CreateTargetGroup'
            Resource: '*'
            Condition:
              'Null':
                'aws:RequestTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:CreateListener'
              - 'elasticloadbalancing:DeleteListener'
              - 'elasticloadbalancing:CreateRule'
              - 'elasticloadbalancing:DeleteRule'
            Resource: '*'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:AddTags'
              - 'elasticloadbalancing:RemoveTags'
            Resource:
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:targetgroup/*/*'
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:loadbalancer/net/*/*'
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:loadbalancer/app/*/*'
            Condition:
              'Null':
                'aws:RequestTag/elbv2.k8s.aws/cluster': 'true'
                'aws:ResourceTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:AddTags'
              - 'elasticloadbalancing:RemoveTags'
            Resource:
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:listener/net/*/*/*'
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:listener/app/*/*/*'
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:listener-rule/net/*/*/*'
              - 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:listener-rule/app/*/*/*'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:ModifyLoadBalancerAttributes'
              - 'elasticloadbalancing:SetIpAddressType'
              - 'elasticloadbalancing:SetSecurityGroups'
              - 'elasticloadbalancing:SetSubnets'
              - 'elasticloadbalancing:DeleteLoadBalancer'
              - 'elasticloadbalancing:ModifyTargetGroup'
              - 'elasticloadbalancing:ModifyTargetGroupAttributes'
              - 'elasticloadbalancing:DeleteTargetGroup'
            Resource: '*'
            Condition:
              'Null':
                'aws:ResourceTag/elbv2.k8s.aws/cluster': 'false'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:RegisterTargets'
              - 'elasticloadbalancing:DeregisterTargets'
            Resource: 'arn:{{ .IAMPolicies.RegionARN }}:elasticloadbalancing:*:*:targetgroup/*/*'
          - Effect: Allow
            Action:
              - 'elasticloadbalancing:SetWebAcl'
              - 'elasticloadbalancing:ModifyListener'
              - 'elasticloadbalancing:AddListenerCertificates'
              - 'elasticloadbalancing:RemoveListenerCertificates'
              - 'elasticloadbalancing:ModifyRule'
            Resource: '*'
{{- if .IAMPolicies.Route53Enabled }}
  Route53ManagerRole:
    Type: "AWS::IAM::Role"
    Properties:
      RoleName: {{ .IAMPolicies.ClusterID }}-Route53Manager-Role
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: "Allow"
            Principal:
              AWS: !GetAtt IAMManagerRole.Arn
            Action: "sts:AssumeRole"
          {{- if ne .IAMPolicies.CloudfrontDomain "" }}
          - Effect: "Allow"
            Principal:
              Federated: "arn:{{ .IAMPolicies.RegionARN }}:iam::{{ .IAMPolicies.AccountID }}:oidc-provider/{{ .IAMPolicies.CloudfrontDomain }}"
            Action: "sts:AssumeRoleWithWebIdentity"
            Condition:
              StringEquals:
                "{{ .IAMPolicies.CloudfrontDomain }}:sub": "system:serviceaccount:kube-system:external-dns"
          {{- end }}
          {{- if ne .IAMPolicies.CloudfrontAliasDomain "" }}
          - Effect: "Allow"
            Principal:
              Federated: "arn:{{ .IAMPolicies.RegionARN }}:iam::{{ .IAMPolicies.AccountID }}:oidc-provider/{{ .IAMPolicies.CloudfrontAliasDomain }}"
            Action: "sts:AssumeRoleWithWebIdentity"
            Condition:
              StringEquals:
                "{{ .IAMPolicies.CloudfrontAliasDomain }}:sub": "system:serviceaccount:kube-system:external-dns"
          {{- end }}
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
