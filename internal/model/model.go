package model

type AgentMetric struct {
	Type  string
	Name  string
	Value any
}

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Error struct {
	Error string `json:"error"`
}

type Data map[string]map[string]Metric
