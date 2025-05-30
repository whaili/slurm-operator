{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Define compute name
*/}}
{{- define "slurm.compute.name" -}}
{{- printf "%s-compute" (include "slurm.fullname" .) -}}
{{- end }}

{{/*
Define compute port
*/}}
{{- define "slurm.compute.port" -}}
{{- print "6818" -}}
{{- end }}

{{/*
Determine compute extraConf (e.g. `--conf <extraConf>`)
*/}}
{{- define "slurm.compute.extraConf" -}}
{{- $extraConf := list -}}
{{- if .extraConf -}}
  {{- $extraConf = splitList " " .extraConf -}}
{{- else if .extraConfMap -}}
  {{- $extraConf = (include "_toList" .extraConfMap) | splitList ";" -}}
{{- end }}
{{- join " " $extraConf -}}
{{- end }}

{{/*
Determine compute partition config
*/}}
{{- define "slurm.compute.PartitionConfig" -}}
{{- $config := list -}}
{{- if .config -}}
  {{- $config = list .config -}}
{{- else if .configMap -}}
  {{- $config = (include "_toList" .configMap) | splitList ";" -}}
{{- end }}
{{- join " " $config -}}
{{- end }}

{{/*
Returns the parsed resource limits for POD_CPUS.
*/}}
{{- define "slurm.compute.podCpus" -}}
{{- $out := 0 -}}
{{- with .resources }}{{- with .limits }}{{- with .cpu }}
  {{- $out = include "resource-quantity" . | float64 | ceil | int -}}
{{- end }}{{- end }}{{- end }}
{{- print $out -}}
{{- end -}}

{{/*
Returns the parsed resource limits for POD_MEMORY, in Megabytes.
*/}}
{{- define "slurm.compute.podMemory" -}}
{{- $out := 0 -}}
{{- with .resources }}{{- with .limits }}{{- with .memory }}
  {{- $megabytes := (include "resource-quantity" "1M") | float64 -}}
  {{- $out = divf (include "resource-quantity" . | float64) $megabytes | ceil | int -}}
{{- end }}{{- end }}{{- end }}
{{- print $out -}}
{{- end -}}
