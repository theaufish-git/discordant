## container.secrets.env secrets
{{- define "container.secrets.env" -}}
{{- $container := index . 0 -}}
{{- $allSecrets := index . 1 -}}
{{- $secrets := index $allSecrets $container -}}
{{- range $secret := $secrets }}
{{- range $key, $value := $secret.values }}
- name: {{ if $secret.prefix }}{{ upper $secret.prefix }}_{{ end }}{{ upper $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $secret.from }}
      key: {{ $value }}
{{- end }}
{{- end }}
{{- end -}}

## container.secrets.mounts container secretFiles
{{- define "container.secrets.mounts" -}}
{{- $container := index . 0 -}}
{{- $allSecretFiles := index . 1 -}}
{{- $secretFiles := index $allSecretFiles $container -}}
{{- range $secretFile := $secretFiles }}
- name: {{ $secretFile.from }}
  mountPath: {{ $secretFile.path }}
  readOnly: true
{{- end }}
{{- end -}}


## container.secrets.volumes container secretFiles
{{- define "container.secrets.volumes" -}}
{{- $container := index . 0 -}}
{{- $allSecretFiles := index . 1 -}}
{{- $secretFiles := index $allSecretFiles $container -}}
{{- range $secretFile := $secretFiles }}
- name: {{ $secretFile.from }}
  secret:
    secretName: {{ $secretFile.from }}
{{- end }}
{{- end -}}
