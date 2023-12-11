package model

type Metric struct {
	Type  string
	Name  string
	Value any
}

type Data map[string]map[string]Metric

const HTMLTemplateString = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <ul>
    {{range .}}{{range .}}
        <li>{{.Name}}: {{.Value}}</li>
    {{end}}{{end}}
    </ul>
</body>
</html>
`
