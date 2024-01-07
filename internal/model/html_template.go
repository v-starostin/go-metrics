package model

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
        <li>{{.ID}}: {{.Value}}{{.Delta}}</li>
    {{end}}{{end}}
    </ul>
</body>
</html>
`
