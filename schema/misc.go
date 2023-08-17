package schema

type ValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type GroupValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Group string `json:"group"`
}

type KV struct {
	Key   string `json:"value"`
	Value string `json:"label"`
	Group string `json:"group"`
}
