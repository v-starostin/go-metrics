package model

type AgentMetric struct {
	MType string `json:"type"`
	ID    string `json:"id"`
	Value any    `json:"value,omitempty"`
	Delta any    `json:"delta,omitempty"`
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

//easyjson:json
type AgentMetrics []AgentMetric
