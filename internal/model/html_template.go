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
        <li>{{.Name}}: {{.Value}}</li>
		<li>{{.Name}}: {{.Delta}}</li>
    {{end}}{{end}}
    </ul>
</body>
</html>
`
