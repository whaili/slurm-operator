{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Determine accounting extraConf
*/}}
{{- define "slurm.accounting.extraConf" -}}
{{- $extraConf := list -}}
{{- if .Values.accounting.extraConf -}}
  {{- $extraConf = splitList "\n" .Values.accounting.extraConf -}}
{{- else if .Values.accounting.extraConfMap -}}
  {{- $extraConf = (include "_toList" .Values.accounting.extraConfMap) | splitList ";" -}}
{{- end }}
{{- join "\n" $extraConf -}}
{{- end }}

{{/*
Define slurm accounting storageConfig
*/}}
{{- define "slurm.accounting.storageConfig" -}}
{{- $storageConfig := dict "passwordKeyRef" dict -}}
{{- if and .Values.accounting.enabled .Values.mariadb.enabled -}}
  {{- $storageHost := include "slurm.accounting.storageHost" . -}}
  {{- $_ := set $storageConfig "host" $storageHost -}}
  {{- $_ := set $storageConfig "database" .Values.mariadb.auth.database -}}
  {{- $_ := set $storageConfig "username" .Values.mariadb.auth.username -}}
  {{- $secretName := include "slurm.accounting.secretName" . -}}
  {{- $_ := set $storageConfig.passwordKeyRef "name" $secretName -}}
  {{- $_ := set $storageConfig.passwordKeyRef "key" "mariadb-password" -}}
{{- else if .Values.accounting.enabled }}
{{- $storageConfig = .Values.accounting.storageConfig -}}
{{- end }}
{{- toYaml $storageConfig }}
{{- end }}

{{- define "slurm.accounting.secretName" -}}
{{- template "mariadb.secretName" .Subcharts.mariadb }}
{{- end }}

{{- define "slurm.accounting.storageHost" -}}
{{- template "mariadb.primary.fullname" .Subcharts.mariadb }}
{{- end }}
