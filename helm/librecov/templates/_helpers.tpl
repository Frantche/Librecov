{{/*
Expand the name of the chart.
*/}}
{{- define "librecov.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "librecov.fullname" -}}
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
{{- define "librecov.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "librecov.labels" -}}
helm.sh/chart: {{ include "librecov.chart" . }}
{{ include "librecov.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "librecov.selectorLabels" -}}
app.kubernetes.io/name: {{ include "librecov.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "librecov.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "librecov.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database host
*/}}
{{- define "librecov.databaseHost" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "librecov.fullname" .) }}
{{- else }}
{{- .Values.externalDatabase.host }}
{{- end }}
{{- end }}

{{/*
Database port
*/}}
{{- define "librecov.databasePort" -}}
{{- if .Values.postgresql.enabled }}
{{- print "5432" }}
{{- else }}
{{- .Values.externalDatabase.port }}
{{- end }}
{{- end }}

{{/*
Database name
*/}}
{{- define "librecov.databaseName" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.externalDatabase.database }}
{{- end }}
{{- end }}

{{/*
Database user
*/}}
{{- define "librecov.databaseUser" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.username }}
{{- else }}
{{- .Values.externalDatabase.user }}
{{- end }}
{{- end }}

{{/*
Database password secret name
*/}}
{{- define "librecov.databaseSecretName" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "librecov.fullname" .) }}
{{- else if .Values.externalDatabase.passwordExistingSecretName }}
{{- .Values.externalDatabase.passwordExistingSecretName }}
{{- else }}
{{- printf "%s-database" (include "librecov.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Database password secret key
*/}}
{{- define "librecov.databaseSecretKey" -}}
{{- if .Values.postgresql.enabled }}
{{- "password" }}
{{- else if .Values.externalDatabase.passwordExistingSecretName }}
{{- .Values.externalDatabase.passwordExistingSecretKey }}
{{- else }}
{{- "db-password" }}
{{- end }}
{{- end }}
