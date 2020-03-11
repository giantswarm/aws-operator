{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "aws-operator.name" -}}
{{- default .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "aws-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "aws-operator.labels" -}}
helm.sh/chart: {{ include "aws-operator.chart" . }}
{{ include "aws-operator.selectorLabels" . }}
app.giantswarm.io/branch: {{ .Values.project.branch }}
app.giantswarm.io/commit: {{ .Values.project.commit }}
app.kubernetes.io/name: {{ include "aws-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "aws-operator.selectorLabels" -}}
app: {{ include "aws-operator.name" . }}
version: {{ .Chart.Version }}
app.giantswarm.io/version: {{ .Chart.AppVersion }}
{{- end -}}
