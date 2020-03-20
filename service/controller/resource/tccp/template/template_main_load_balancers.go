package template

const TemplateMainLoadBalancers = `
{{- define "load_balancers" -}}
{{- $v := .LoadBalancers }}
  ApiInternalLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      ConnectionSettings:
        IdleTimeout: 1200
      HealthCheck:
        HealthyThreshold: 2
        Interval: 5
        Target: {{ $v.APIElbHealthCheckTarget }}
        Timeout: 3
        UnhealthyThreshold: 2
      Instances:
      - !Ref {{ $v.MasterInstanceResourceName }}
      Listeners:
      {{ range $v.APIElbPortsToOpen}}
      - InstancePort: {{ .PortInstance }}
        InstanceProtocol: TCP
        LoadBalancerPort: {{ .PortELB }}
        Protocol: TCP
      {{ end }}
      LoadBalancerName: {{ $v.APIInternalElbName }}
      Scheme: internal
      SecurityGroups:
        - !Ref APIInternalELBSecurityGroup
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{end}}
  ApiLoadBalancer:
    Type: AWS::ElasticLoadBalancing::LoadBalancer
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      ConnectionSettings:
        IdleTimeout: 1200
      HealthCheck:
        HealthyThreshold: 2
        Interval: 5
        Target: {{ $v.APIElbHealthCheckTarget }}
        Timeout: 3
        UnhealthyThreshold: 2
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
      Scheme: internet-facing
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
        HealthyThreshold: 2
        Interval: 5
        Target: {{ $v.EtcdElbHealthCheckTarget }}
        Timeout: 3
        UnhealthyThreshold: 2
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
      Scheme: internal
      SecurityGroups:
        - !Ref EtcdELBSecurityGroup
      Subnets:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{end}}
{{- end -}}
`
