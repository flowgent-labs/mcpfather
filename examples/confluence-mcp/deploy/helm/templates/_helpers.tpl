{{/*
Expand the name of the chart.
*/}}
{{- define "confluence-mcp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "confluence-mcp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "confluence-mcp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "confluence-mcp.labels" -}}
helm.sh/chart: {{ include "confluence-mcp.chart" . | quote }}
{{ include "confluence-mcp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/part-of: mcpfather
{{- end }}

{{/*
Selector labels
*/}}
{{- define "confluence-mcp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "confluence-mcp.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "confluence-mcp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "confluence-mcp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Secret name to use
*/}}
{{- define "confluence-mcp.secretName" -}}
{{- if eq .Values.secret.provider "static" }}
{{- if .Values.secret.static.create }}
{{- include "confluence-mcp.fullname" . }}-secret
{{- else }}
{{- required "secret.static.externalName is required when static.create is false and provider is static" .Values.secret.static.externalName }}
{{- end }}
{{- else if eq .Values.secret.provider "gcp" }}
{{- include "confluence-mcp.fullname" . }}-gcp-secret
{{- else }}
{{- required "secret.provider must be one of: static, gcp" .Values.secret.provider }}
{{- end }}
{{- end }}

{{/*
Return "true" when static k8s Secret should be rendered, else "".
*/}}
{{- define "confluence-mcp.useStaticSecret" -}}
{{- if and (eq .Values.secret.provider "static") .Values.secret.static.create }}true{{- end }}
{{- end }}

{{/*
Return "true" when GCP Secret Manager is configured, else "".
*/}}
{{- define "confluence-mcp.useGCPSecret" -}}
{{- if eq .Values.secret.provider "gcp" }}true{{- end }}
{{- end }}

{{/*
Return "true" when Envoy Gateway API should be used, else "".
*/}}
{{- define "confluence-mcp.useEnvoyGateway" -}}
{{- if .Values.ingress.envoy.enabled }}true{{- end }}
{{- end }}

{{/*
Return "true" when standard Ingress (nginx) should be used, else "".
*/}}
{{- define "confluence-mcp.useNginxIngress" -}}
{{- if .Values.ingress.nginx.enabled }}true{{- end }}
{{- end }}
