{{ if .items }}{{ range .items -}}
{{ if not .status.redeemedBy -}}
http://localhost:4200/invitations/{{ .metadata.name }}?token={{ .status.token }}
{{ end -}}
{{ end -}}
{{ else -}}
http://localhost:4200/invitations/{{ .metadata.name }}?token={{ .status.token }}
{{ end -}}
