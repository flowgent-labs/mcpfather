{{/*
Expand the name of the chart.
*/}}
{{- define "__BINARY_NAME__.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "__BINARY_NAME__.fullname" -}}
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
{{- define "__BINARY_NAME__.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "__BINARY_NAME__.labels" -}}
helm.sh/chart: {{ include "__BINARY_NAME__.chart" . | quote }}
{{ include "__BINARY_NAME__.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/part-of: mcpfather
{{- end }}

{{/*
Selector labels
*/}}
{{- define "__BINARY_NAME__.selectorLabels" -}}
app.kubernetes.io/name: {{ include "__BINARY_NAME__.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "__BINARY_NAME__.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "__BINARY_NAME__.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Secret name to use
*/}}
{{- define "__BINARY_NAME__.secretName" -}}
{{- if eq .Values.secret.provider "static" }}
{{- if .Values.secret.static.create }}
{{- include "__BINARY_NAME__.fullname" . }}-secret
{{- else }}
{{- required "secret.static.externalName is required when static.create is false and provider is static" .Values.secret.static.externalName }}
{{- end }}
{{- else if eq .Values.secret.provider "gcp" }}
{{- include "__BINARY_NAME__.fullname" . }}-gcp-secret
{{- else }}
{{- required "secret.provider must be one of: static, gcp" .Values.secret.provider }}
{{- end }}
{{- end }}

{{/*
Return "true" when static k8s Secret should be rendered, else "".
*/}}
{{- define "__BINARY_NAME__.useStaticSecret" -}}
{{- if and (eq .Values.secret.provider "static") .Values.secret.static.create }}true{{- end }}
{{- end }}

{{/*
Return "true" when GCP Secret Manager is configured, else "".
*/}}
{{- define "__BINARY_NAME__.useGCPSecret" -}}
{{- if eq .Values.secret.provider "gcp" }}true{{- end }}
{{- end }}

{{/*
Return "true" when Envoy Gateway API should be used, else "".
*/}}
{{- define "__BINARY_NAME__.useEnvoyGateway" -}}
{{- if .Values.ingress.envoy.enabled }}true{{- end }}
{{- end }}

{{/*
Return "true" when standard Ingress (nginx) should be used, else "".
*/}}
{{- define "__BINARY_NAME__.useNginxIngress" -}}
{{- if .Values.ingress.nginx.enabled }}true{{- end }}
{{- end }}
