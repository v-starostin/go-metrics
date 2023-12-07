package model

var GaugeMetrics = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
	"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
	"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
	"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
}

type Metric struct {
	Type  string
	Name  string
	Value any
}

type Data map[string]map[string]Metric

type Error struct {
	Error string
}

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
