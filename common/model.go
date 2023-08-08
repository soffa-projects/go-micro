package common

type IntValue struct {
	Value int `json:"value"`
}

func (v *IntValue) Inc() {
	v.Value++
}
