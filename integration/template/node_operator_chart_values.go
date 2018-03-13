package template

// NodeOperatorChartValues values required by node-operator-chart, the
// environment variables will be expanded before writing the contents to a file.
var NodeOperatorChartValues = `Installation:
  V1:
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
