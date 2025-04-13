package models

type NewProxySettings struct {
	Name      string
	Namespace string
	Version   string
	Secret    string
	Records   map[string]string
}
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
type Health struct {
	Status  string `json:"status,omitempty"`
	Time    string `json:"time,omitempty"`
	Version string `json:"version,omitempty"`
}
