{{- define "mycelis-core.coreAuthSecretName" -}}
{{- if .Values.coreAuth.secretName }}
{{- .Values.coreAuth.secretName -}}
{{- else -}}
{{ printf "%s-core-auth" .Chart.Name }}
{{- end -}}
{{- end -}}
