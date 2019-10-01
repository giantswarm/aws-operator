package template

const TemplateMainLoadBalancers = `
{{- define "load_balancers" -}}
{{- $v := .Guest.LoadBalancers }}
  ApiInternalLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      Name: {{ $v.APIInternalElbName }}
      Scheme: {{ $v.APIInternalElbScheme }}
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{- end}}
      Type: network
  {{- range $v.APIElbListenersAndTargets}}
  {{ .TargetResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      HealthCheckEnabled: true
      HealthCheckIntervalSeconds: {{ $v.ELBHealthCheckInterval }}
      HealthCheckPort: {{ .PortInstance }}
      HealthCheckProtocol: TCP
      HealthyThresholdCount: {{ $v.ELBHealthCheckHealthyThreshold }}
      Port: {{ .PortInstance }}
      Protocol: TCP
      Targets:
      - Id: !Ref {{ $v.MasterInstanceResourceName }}
      TargetType: instance
      UnhealthyThresholdCount: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      VpcId: !Ref VPC
  {{ .ListenerResourceName }}:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
      - Type: forward
        TargetGroupArn: !Ref {{ .TargetResourceName }}
      LoadBalancerArn: !Ref ApiInternalLoadBalancer
      Port: {{ .PortELB }}
      Protocol: TCP
  {{- end }}

  ApiLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      ConnectionSettings:
        IdleTimeout: 1200
      HealthCheck:
        HealthyThreshold: {{ $v.ELBHealthCheckHealthyThreshold }}
        Interval: {{ $v.ELBHealthCheckInterval }}
        Target: {{ $v.APIElbHealthCheckTarget }}
        Timeout: {{ $v.ELBHealthCheckTimeout }}
        UnhealthyThreshold: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      Instances:
      - !Ref {{ $v.MasterInstanceResourceName }}
      Listeners:
      {{ range $v.APIElbListenersAndTargets}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ $v.APIElbName }}
      Scheme: {{ $v.APIElbScheme }}
      SecurityGroups:
        - !Ref MasterSecurityGroup
      Subnets:
      {{- range $s := $v.PublicSubnets }}
        - !Ref {{ $s }}
      {{end}}

  EtcdLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    Properties:
      ConnectionSettings:
        IdleTimeout: 1200
      HealthCheck:
        HealthyThreshold: {{ $v.ELBHealthCheckHealthyThreshold }}
        Interval: {{ $v.ELBHealthCheckInterval }}
        Target: {{ $v.EtcdElbHealthCheckTarget }}
        Timeout: {{ $v.ELBHealthCheckTimeout }}
        UnhealthyThreshold: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      Instances:
      - !Ref {{ $v.MasterInstanceResourceName }}
      Listeners:
      {{ range $v.EtcdElbListenersAndTargets}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ $v.EtcdElbName }}
      Scheme: {{ $v.EtcdElbScheme }}
      SecurityGroups:
        - !Ref EtcdELBSecurityGroup
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{end}}

  IngressLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      ConnectionSettings:
        IdleTimeout: 60
      HealthCheck:
        HealthyThreshold: {{ $v.ELBHealthCheckHealthyThreshold }}
        Interval: {{ $v.ELBHealthCheckInterval }}
        Target: {{ $v.IngressElbHealthCheckTarget }}
        Timeout: {{ $v.ELBHealthCheckTimeout }}
        UnhealthyThreshold: {{ $v.ELBHealthCheckUnhealthyThreshold }}
      Listeners:
      {{ range $v.IngressElbListenersAndTargets}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ $v.IngressElbName }}
      Policies:
      - PolicyName: "EnableProxyProtocol"
        PolicyType: "ProxyProtocolPolicyType"
        Attributes:
        - Name: "ProxyProtocol"
          Value: "true"
        InstancePorts:
        {{ range $v.IngressElbListenersAndTargets}}
        - {{ .PortInstance }}
        {{ end }}
      Scheme: {{ $v.IngressElbScheme }}
      SecurityGroups:
        - !Ref IngressSecurityGroup
      Subnets:
      {{- range $s := $v.PublicSubnets }}
        - !Ref {{ $s }}
      {{end}}
{{- end -}}
`
