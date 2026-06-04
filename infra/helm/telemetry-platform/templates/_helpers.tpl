{{/*
Labels padrão Helm aplicados em todos os recursos.
*/}}
{{- define "telemetry-platform.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/part-of: {{ .Chart.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Constrói a referência de imagem respeitando o registry global.
Uso: {{ include "telemetry-platform.image" (dict "registry" .Values.global.imageRegistry "name" "telemetry/ingestion-service" "tag" .Values.global.imageTag) }}
*/}}
{{- define "telemetry-platform.image" -}}
{{- if .registry -}}
{{ .registry }}/{{ .name }}:{{ .tag }}
{{- else -}}
{{ .name }}:{{ .tag }}
{{- end -}}
{{- end }}
