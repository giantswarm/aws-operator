package template

const TemplateMainSecurityGroups = `
{{- define "security_groups" -}}
{{- $v := .SecurityGroups -}}
  MasterSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-master
      VpcId: !Ref VPC
      SecurityGroupIngress:
      -
        Description: "Allow traffic from Control Plane CIDR to 4194 for cadvisor scraping."
        Port: 4194
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Control Plane CIDR to 2379 for etcd backup."
        Port: 2379
        Protocol: tcp
        sourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Control Plane CIDR to 10250 for kubelet scraping."
        Port: 10250
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Control Plane CIDR to 10300 for node-exporter scraping."
        Port: 10300
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Control Plane CIDR to 10301 for kube-state-metrics scraping."
        Port: 10301
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Only allow SSH traffic from the Control Plane."
        Port: 22
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR

      #
      # Public API Whitelist Enabled Rules
      #
      {{- if $v.APIWhitelist.Public.Enabled }}
      -
        Description: "Allow traffic from Control Plane CIDR."
        Port: 443
        Protocol: tcp
        sourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Tenant Cluster CIDR."
        Port: 443
        Protocol: tcp
        SourceCIDR: $v.TenantClusterVPCCIDR

      {{- range $subnet := $v.APIWhitelist.Public.SubnetList }}
      -
        Description: "Custom Public API Whitelist CIDR."
        Port: 443
        Protocol: tcp
        SourceCIDR: $subnet
      {{- end }}

      {{- range $v.ControlPlaneNATGatewayAddresses }}
      -
        Description: "Allow traffic from NAT Gateways."
        Port: 443
        Protocol: tcp
        SourceCIDR: {{ .PublicIp }}/32
      {{- end }}

      {{- range .NATGateway.Gateways }}
      -
        Description: "Allow NAT gateway IP."
        IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: !Join [ "/", [ !Ref {{ .NATEIPName }}, "32" ] ]
      {{- end }}

      #
      # Public API Whitelist Disabled Rules
      #
      {{- else }}
      -
        Description: "Allow all traffic to the master instance."
        Port: 443
        Protocol: tcp
        SourceCIDR: 0.0.0.0/0
      {{- end }}

      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-master
  EtcdELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-etcd-elb
      VpcId: !Ref VPC
      SecurityGroupIngress:
      -
        Description: "Allow all Etcd traffic from the VPC to the Etcd load balancer."
        Port: 2379
        Protocol: tcp
        SourceCIDR: 0.0.0.0/0
      -
        Description: "Allow traffic from Control Plane to Etcd port for backup and metrics."
        Port: 2379
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-etcd-elb
  APIInternalELBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: {{ $v.ClusterID }}-internal-api
      VpcId: !Ref VPC
      SecurityGroupIngress:

      #
      # Private API Whitelist Enabled Rules
      #
      {{- if $v.APIWhitelist.Private.Enabled }}
      -
        Description: "Allow traffic from Control Plane CIDR."
        Port: 443
        Protocol: tcp
        SourceCIDR: $v.ControlPlaneVPCCIDR
      -
        Description: "Allow traffic from Tenant Cluster CIDR."
        Port: 443
        Protocol: tcp
        SourceCIDR: $v.TenantClusterVPCCIDR

      {{- range $subnet := $v.APIWhitelist.Private.SubnetList }}
      -
        Description: "Custom Private API Whitelist CIDR."
        Port: 443
        Protocol: tcp
        SourceCIDR: $subnet
      {{- end }}

      #
      # Private API Whitelist Disabled Rules
      #
      {{- else }}
      -
        Description: "Allow all traffic to the master instance from A class network."
        Port: 443
        Protocol: tcp
        SourceCIDR: "10.0.0.0/8"
      -
        Description: "Allow all traffic to the master instance from B class network."
        Port: 443
        Protocol: tcp
        SourceCIDR: "172.16.0.0/12"
      -
        Description: "Allow all traffic to the master instance from C class network."
        Port: 443
        Protocol: tcp
        SourceCIDR: "192.168.0.0/16"
      {{- end }}

      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-internal-api
  AWSCNISecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: "AWS CNI Security Group configured to the ENIConfig CRD."
      VpcId: !Ref VPC
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-aws-cni
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
  MasterAllowCalicoIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: -1
      FromPort: -1
      ToPort: -1
      SourceSecurityGroupId: !Ref MasterSecurityGroup
  MasterAllowPodsCNIIngressRule:
      Type: AWS::EC2::SecurityGroupIngress
      DependsOn: MasterSecurityGroup
      Properties:
        Description: Allow traffic from pod to master.
        GroupId: !Ref MasterSecurityGroup
        IpProtocol: -1
        FromPort: -1
        ToPort: -1
        SourceSecurityGroupId: !Ref AWSCNISecurityGroup
  MasterAllowEtcdIngressRule:
    Type: AWS::EC2::SecurityGroupIngress
    DependsOn: MasterSecurityGroup
    Properties:
      GroupId: !Ref MasterSecurityGroup
      IpProtocol: "tcp"
      FromPort: 2379
      ToPort: 2379
      SourceSecurityGroupId: !Ref EtcdELBSecurityGroup
  VPCDefaultSecurityGroupEgress:
    Type: AWS::EC2::SecurityGroupEgress
    Properties:
      Description: Allow outbound traffic from loopback address.
      GroupId: !GetAtt VPC.DefaultSecurityGroup
      IpProtocol: -1
      CidrIp: 127.0.0.1/32
{{- end -}}
`
