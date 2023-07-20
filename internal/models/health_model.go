package models

type SystemHealth struct {
	Version        string         `json:"version"`
	ServiceSupport ServiceSupport `json:"service_support"`
}

type ServiceSupport struct {
	Master bool `json:"master"`
	Slave  bool `json:"slave"`
	Redis  bool `json:"redis"`
}
