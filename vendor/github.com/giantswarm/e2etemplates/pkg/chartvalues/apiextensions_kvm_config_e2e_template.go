package chartvalues

var apiExtensionsKVMConfigE2ETemplate = `
baseDomain: "k8s.gastropod.gridscale.kvm.gigantic.io"
cluster:
  id: "{{.ClusterID}}"
encryptionKey: "QitRZGlWeW5WOFo2YmdvMVRwQUQ2UWoxRHZSVEF4MmovajlFb05sT1AzOD0="
kvm:
  vni: {{.VNI}}
  ingress:
    httpNodePort: {{.HttpNodePort}}
    httpTargetPort: 30010
    httpsNodePort: {{.HttpsNodePort}}
    httpsTargetPort: 30011
sshUser: "test-user"
sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAYQCurvzg5Ia54kb3NZapA6yP00//+Jt6XJNeC7Seq3TeCqMR9x7Snalj19r0lWok1PkRgDo1PXj+3y53zo/wqBrPqN4cQqp00R06kNfnhAgesaRMvYhuyVRQQbfXV5gQg8M= dummy-key"
versionBundle:
  version: "{{.VersionBundleVersion}}"
`
