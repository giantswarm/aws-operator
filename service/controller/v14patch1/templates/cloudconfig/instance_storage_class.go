package cloudconfig

const InstanceStorageClass = `
write_files:
- path: /srv/default-storage-class.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: storage.k8s.io/v1beta1
    kind: StorageClass
    metadata:
      name: gp2
      annotations:
        storageclass.beta.kubernetes.io/is-default-class: "true"
      labels:
        kubernetes.io/cluster-service: "true"
        addonmanager.kubernetes.io/mode: EnsureExists
    provisioner: kubernetes.io/aws-ebs
    allowVolumeExpansion: true
    parameters:
      type: gp2
`
const InstanceStorageClassEncrypted = InstanceStorageClass + `
      encrypted: "true"
`
