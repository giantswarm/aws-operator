package chartvalues

const releaseOperatorTemplate = `Installation:
  V1:
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
`
