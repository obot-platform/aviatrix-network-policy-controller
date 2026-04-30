{{- define "aviatrix-network-policy-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aviatrix-network-policy-controller.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "aviatrix-network-policy-controller.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "aviatrix-network-policy-controller.serviceAccountName" -}}
{{- if .Values.serviceAccount.name -}}
{{- .Values.serviceAccount.name -}}
{{- else -}}
{{- include "aviatrix-network-policy-controller.fullname" . -}}
{{- end -}}
{{- end -}}

{{- define "aviatrix-network-policy-controller.obotRoleName" -}}
{{- printf "%s-obot" (include "aviatrix-network-policy-controller.fullname" . | trunc 58 | trimSuffix "-") | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Get the image tag, defaulting to appVersion. If appVersion looks like a development version,
defaults to the "main" tag instead.
*/}}
{{- define "aviatrix-network-policy-controller.imageTag" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if and (not .Values.image.tag) (hasPrefix "0.0.0" .Chart.AppVersion) -}}
{{- $tag = "main" -}}
{{- end -}}
{{- $tag -}}
{{- end -}}
