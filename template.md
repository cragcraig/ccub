# {{.Date}}  {{.Title}}  -  {{.Assembly}}
{{range .WorkPeriod}}
  * {{.StartTime}}-{{.EndTime}} ({{.DurationMin}} minutes)
{{end}}
{{.Details}}

