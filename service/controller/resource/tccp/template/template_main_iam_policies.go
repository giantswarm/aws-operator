package template

const TemplateMainIAMPolicies = `
{{- define "iam_policies" -}}
{{- $v := .IAMPolicies -}}
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
{{- end -}}
`
