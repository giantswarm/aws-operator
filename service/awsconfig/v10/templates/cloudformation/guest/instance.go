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
      Volumes:
      - Device: /dev/sdh
        VolumeId: !Ref EtcdVolume
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
        Value: {{ .Instance.Cluster.ID }}-etcd
{{end}}`
