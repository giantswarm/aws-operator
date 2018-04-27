package e2etemplates

// NodeOperatorChartValues values required by node-operator-chart, the
// environment variables will be expanded before writing the contents to a file.
const NodeOperatorChartValues = `Installation:
  V1:
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
`
