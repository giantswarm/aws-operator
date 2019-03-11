package cloudconfig

const InstanceStorageClassContent = `apiVersion: storage.k8s.io/v1beta1
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
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: gp2
`
const InstanceStorageClassEncryptedContent = InstanceStorageClassContent + `
  encrypted: "true"
`
