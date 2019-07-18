package tccp

const InternetGateway = `
{{- define "internet_gateway" -}}
{{- $v := .Guest.InternetGateway -}}
  InternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}
        - Key: giantswarm.io/tccp
          Value: true
  VPCGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    DependsOn:
      - {{ $v.PublicRouteTableName }}
      {{- range $rt := $v.PrivateRouteTables }}
      - {{ $rt }}
      {{- end }}
    Properties:
      InternetGatewayId:
        Ref: InternetGateway
      VpcId: !Ref VPC
  InternetGatewayRoute:
    Type: AWS::EC2::Route
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      RouteTableId: !Ref {{ $v.PublicRouteTableName }}
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId:
        Ref: InternetGateway
{{- end -}}
`
