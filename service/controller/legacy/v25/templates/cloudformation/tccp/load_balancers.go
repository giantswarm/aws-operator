package tccp

const LoadBalancers = `
{{define "load_balancers"}}
{{- $v := .Guest.LoadBalancers }}
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
      {{ range $v.APIElbPortsToOpen}}
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
      {{ range $v.EtcdElbPortsToOpen}}
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
      {{ range $v.IngressElbPortsToOpen}}
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
        {{ range $v.IngressElbPortsToOpen}}
        - {{ .PortInstance }}
        {{ end }}
      Scheme: {{ $v.IngressElbScheme }}
      SecurityGroups:
        - !Ref IngressSecurityGroup
      Subnets:
      {{- range $s := $v.PublicSubnets }}
        - !Ref {{ $s }}
      {{end}}
{{end}}
`
