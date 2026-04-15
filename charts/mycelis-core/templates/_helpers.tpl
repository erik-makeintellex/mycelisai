{{- define "mycelis-core.coreAuthSecretName" -}}
{{- if .Values.coreAuth.secretName }}
{{- .Values.coreAuth.secretName -}}
{{- else -}}
{{ printf "%s-core-auth" .Chart.Name }}
{{- end -}}
{{- end -}}

{{- define "mycelis-core.image" -}}
{{- if .Values.image.digest -}}
{{ printf "%s@%s" .Values.image.repository .Values.image.digest }}
{{- else -}}
{{ printf "%s:%s" .Values.image.repository (default "dev" .Values.image.tag) }}
{{- end -}}
{{- end -}}

{{- define "mycelis-core.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default .Chart.Name .Values.serviceAccount.name -}}
{{- else -}}
{{- .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
