package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .SecurityGroups -}}
  AWSCNISecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: "AWS CNI Security Group configured to the ENIConfig CRD."
      VpcId: !Ref VPC
      Tags:
        - Key: Name
          Value: {{ $v.AWSCNISecurityGroupName }}
  PodsIngressRuleFromMAsters:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: AWSCNISecurityGroup
    Properties:
      Description: Allow traffic from masters to pods.
      GroupId: !Ref AWSCNISecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  PodsAllowPodsCNIIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: AWSCNISecurityGroup
    Properties:
      Description: Allow traffic from pod to pod.
      GroupId: !Ref AWSCNISecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref AWSCNISecurityGroup
  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      Description: Allow outbound traffic from loopback address.
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      IpProtocol: -1
      CidrIp: 127.0.0.1/32
{{- end -}}
`
