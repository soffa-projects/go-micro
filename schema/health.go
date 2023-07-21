package schema

type ComponentStatus struct {
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

type HealthStatus struct {
	Status     string                     `json:"status"`
	Components map[string]ComponentStatus `json:"components,omitempty"`
}

type ErrorResponse struct {
	Kind    string      `json:"kind"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

func NewHealthStatus() *HealthStatus {
	return &HealthStatus{Status: "UP", Components: make(map[string]ComponentStatus)}
}

func (h *HealthStatus) SetComponentStatus(name string, err error) {
	status := "UP"
	details := ""
	if err != nil {
		status = "DOWN"
		details = err.Error()
		h.Status = "DOWN"
	}
	h.Components[name] = ComponentStatus{
		Status:  status,
		Details: details,
	}
}
