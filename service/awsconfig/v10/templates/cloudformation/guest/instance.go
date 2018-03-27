package guest

const Instance = `{{define "instance"}}
  {{ .Instance.Master.Instance.ID }}:
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
    - {{ .Instance.Master.Instance.ID }}
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: !GetAtt {{ .Instance.Master.Instance.ID }}.AvailabilityZone
      Tags:
      - Key: Name
        Value: {{ .Instance.Cluster.ID }}-etcd
  MountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ .Instance.Master.Instance.ID }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/sdh
{{end}}`
