package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
  GeneralSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: General Node Pool Security Group For Basic Traffic Rules.
      SecurityGroupIngress:
      -
        Description: Allow traffic from control plane CIDR to 22 for SSH access.
        IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from tenant cluster CIDR to 2049 for NFS access.
        IpProtocol: tcp
        FromPort: 2049
        ToPort: 2049
        CidrIp: {{ .SecurityGroups.TenantCluster.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 4194 for cadvisor scraping.
        IpProtocol: tcp
        FromPort: 4194
        ToPort: 4194
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10250 for kubelet scraping.
        IpProtocol: tcp
        FromPort: 10250
        ToPort: 10250
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10300 for node-exporter scraping.
        IpProtocol: tcp
        FromPort: 10300
        ToPort: 10300
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      -
        Description: Allow traffic from control plane CIDR to 10301 for kube-state-metrics scraping.
        IpProtocol: tcp
        FromPort: 10301
        ToPort: 10301
        CidrIp: {{ .SecurityGroups.ControlPlane.VPC.CIDR }}
      Tags:
        - Key: Name
          Value: {{ .SecurityGroups.ClusterID }}-worker
      VpcId: {{ .SecurityGroups.TenantCluster.VPC.ID }}
  GeneralInternalAPIIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from the TCNP General Security Group to the TCCP Internal API Security Group.
      GroupId: {{ .SecurityGroups.TenantCluster.InternalAPI.ID }}
      IpProtocol: tcp
      FromPort: 443
      ToPort: 443
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  GeneralMasterIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from the TCNP General Security Group to the TCCP Master Security Group.
      GroupId: {{ .SecurityGroups.TenantCluster.Master.ID }}
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  {{- if .SecurityGroups.EnableAWSCNI }}
  PodsIngressRuleFromWorkers:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from workers to pods.
      GroupId: {{ .SecurityGroups.TenantCluster.AWSCNI.ID }}
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  PodsIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from pods to the worker nodes.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: {{ .SecurityGroups.TenantCluster.AWSCNI.ID }}
  {{- end }}
  InternalIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic between workloads within the Node Pool.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref GeneralSecurityGroup
  MasterGeneralIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      Description: Allow traffic from the TCCP Master Security Group to the TCNP General Security Group.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: {{ .SecurityGroups.TenantCluster.Master.ID }}
  {{ range .SecurityGroups.TenantCluster.NodePools }}
  NodePoolToNodePoolRule{{ .ResourceName }}:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: GeneralSecurityGroup
    Properties:
      # The rule description is used for identifying the ingress rule. Thus it
      # must not change. Otherwise the tcnpsecuritygroups resource will not be
      # able to properly find the current and desired state of the ingress
      # rules.
      Description: Allow traffic from other Node Pool Security Groups to the Security Group of this Node Pool.
      GroupId: !Ref GeneralSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: {{ .ID }}
  {{- end -}}
{{- end -}}
`
