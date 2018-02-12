package hostpre

const IAMRoles = `
{{define "iam_roles"}}
  PeerRole:
    Type: 'AWS::IAM::Role'
    Properties:
      RoleName: {{ .PeerAccessRoleName }}
      AssumeRolePolicyDocument:
        Statement:
          - Principal:
              AWS: '{{ .GuestAccountID }}'
            Action:
              - 'sts:AssumeRole'
            Effect: Allow
      Path: /
      Policies:
        - PolicyName: root
          PolicyDocument:
            Version: 2012-10-17
            Statement:
              - Effect: Allow
                Action: 'ec2:AcceptVpcPeeringConnection'
                Resource: '*'
{{end}}
`
