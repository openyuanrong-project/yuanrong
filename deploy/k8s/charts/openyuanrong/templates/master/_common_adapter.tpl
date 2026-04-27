{{- define "functionSystem.imagePullSecrets" -}}
{{- range .Values.global.imagePullSecrets }}
- name: {{ . }}
{{- end }}
{{- end -}}