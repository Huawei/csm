{{- define "leader-election" -}}
{{- if gt ( (.Values.global).replicaCount | int ) 1 -}}
- --enable-leader-election=true
- --leader-lease-duration={{ ((.Values.global).leaderElection).leaseDuration | default "8s" }}
- --leader-renew-deadline={{ ((.Values.global).leaderElection).renewDeadline | default "6s" }}
- --leader-retry-period={{ ((.Values.global).leaderElection).retryPeriod | default "2s" }}
{{- else -}}
- --enable-leader-election=false
{{- end -}}
{{- end -}}

{{- define "log" -}}
- --logging-module={{ .module | default "file" }}
- --log-level={{ .level | default "info" }}
- --log-file-size={{ .fileSize | default "20M" }}
- --max-backups={{ .maxBackups | default 9 }}
{{- end -}}