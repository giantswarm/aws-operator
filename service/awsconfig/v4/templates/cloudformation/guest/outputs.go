package guest

const Outputs = `
{{define "outputs"}}
Outputs:
  MasterImageID:
    Value: {{ .MasterImageID }}
  MasterInstanceType:
    Value: {{ .MasterInstanceType }}
  MasterCloudConfigVersion:
    Value: {{ .MasterCloudConfigVersion }}
  WorkerCount:
    Value: {{ .ASGMinSize }}
  WorkerImageID:
    Value: {{ .WorkerImageID }}
  WorkerInstanceType:
    Value: {{ .WorkerInstanceType }}
  WorkerCloudConfigVersion:
    Value: {{ .WorkerCloudConfigVersion }}
{{end}}
`
