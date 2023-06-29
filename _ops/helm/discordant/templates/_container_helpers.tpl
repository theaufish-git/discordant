## container.env.secrets container
{{- define "container.env.secrets" -}}
{{- range $secret := . }}
{{- range $key, $var := $secret.values }}
- name: {{ if $secret.prefix }}{{ upper $secret.prefix }}_{{ end }}{{ upper $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $secret.from }}
      key: {{ $var }}
{{- end }}
{{- end }}
{{- end -}}
