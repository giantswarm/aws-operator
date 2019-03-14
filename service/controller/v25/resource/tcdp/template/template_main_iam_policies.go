package template

const TemplateMainIAMPolicies = `
{{define "iam_policies"}}
  Role:
    Type: AWS::IAM::Role
    Properties:
      RoleName: {{ .IAMPolicies.RoleName }}
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          Effect: "Allow"
          Principal:
            Service: {{ .IAMPolicies.EC2ServiceDomain }}
          Action: "sts:AssumeRole"
  RolePolicy:
    Type: "AWS::IAM::Policy"
    Properties:
      PolicyName: {{ .IAMPolicies.PolicyName }}
      Roles:
        - Ref: "Role"
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
{{ if .IAMPolicies.KMSKeyARN }}
          - Effect: "Allow"
            Action: "kms:Decrypt"
            Resource: "{{ .IAMPolicies.KMSKeyARN }}"
{{ end }}
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
{{ end }}
`
