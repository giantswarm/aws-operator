{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "aws-operator.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "aws-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a name stem for resource names

When pods for deployments are created they have an additional 16 character
suffix appended, e.g. "-957c9d6ff-pkzgw". Given that Kubernetes allows 63
characters for resource names, the stem is truncated to 47 characters to leave
room for such suffix.
*/}}
{{- define "resource.default.name" -}}
{{- .Release.Name | replace "." "-" | trunc 47 | trimSuffix "-" -}}
{{- end -}}

{{- define "resource.psp.name" -}}
{{- include "resource.default.name" . -}}-psp
{{- end -}}

{{- define "resource.pullSecret.name" -}}
{{- include "resource.default.name" . -}}-pull-secret
{{- end -}}

{{- define "resource.default.namespace" -}}
giantswarm
{{- end -}}

{{/*
Common labels
*/}}
{{- define "aws-operator.labels" -}}
app: {{ include "aws-operator.name" . | quote }}
{{ include "aws-operator.selectorLabels" . }}
app.giantswarm.io/branch: {{ .Values.project.branch | quote }}
app.giantswarm.io/commit: {{ .Values.project.commit | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
helm.sh/chart: {{ include "aws-operator.chart" . | quote }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "aws-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "aws-operator.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end -}}
