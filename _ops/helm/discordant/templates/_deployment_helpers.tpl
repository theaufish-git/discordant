{{- define "deployment.imagePullSecrets" -}}
{{- if .Values.imagePullSecrets -}}
imagePullSecrets:
- name: {{ .Values.imagePullSecrets }}
{{- end -}}
{{- end -}}