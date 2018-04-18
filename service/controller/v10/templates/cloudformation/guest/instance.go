package guest

const Instance = `{{define "instance"}}
  {{ .Instance.Master.Instance.ResourceName }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    Properties:
      AvailabilityZone: {{ .Instance.Master.AZ }}
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ .Instance.Image.ID }}
      InstanceType: {{ .Instance.Master.Instance.Type }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref PrivateSubnet
      UserData: {{ .Instance.Master.CloudConfig }}
      Tags:
      - Key: Name
        Value: {{ .Instance.Cluster.ID }}-master
  EtcdVolume:
    Type: AWS::EC2::Volume
    DependsOn:
    - {{ .Instance.Master.Instance.ResourceName }}
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: !GetAtt {{ .Instance.Master.Instance.ResourceName }}.AvailabilityZone
      Tags:
      - Key: Name
        Value: {{ .Instance.Master.EtcdVolume.Name }}
  {{ .Instance.Master.Instance.ResourceName }}MountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instance.Master.Instance.ResourceName }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/sdh
{{end}}`
