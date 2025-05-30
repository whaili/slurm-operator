{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Define token name
*/}}
{{- define "slurm.token.name" -}}
{{ printf "%s-token-create" (include "slurm.fullname" .) }}
{{- end }}

{{/*
Define token labels
*/}}
{{- define "slurm.token.labels" -}}
app.kubernetes.io/component: token
{{ include "slurm.token.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define token selectorLabels
*/}}
{{- define "slurm.token.selectorLabels" -}}
app.kubernetes.io/name: token
app.kubernetes.io/instance: {{ (include "slurm.fullname" .) }}
{{- end }}

{{/*
Define Slurm key secret
*/}}
{{- define "slurm.authSecret" -}}
{{- if .Values.slurmKey }}
  {{- index .Values.slurmKeyRef "name" }}
{{- else }}
  {{- printf "%s-auth-slurm" (include "slurm.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Define JWT HS256 Key secret
*/}}
{{- define "slurm.jwtSecret" -}}
{{- if .Values.jwtHs256Key }}
  {{- index .Values.jwtHs256KeyRef "name" }}
{{- else }}
  {{- printf "%s-auth-jwths256" (include "slurm.fullname" .) }}
{{- end }}
{{- end }}
