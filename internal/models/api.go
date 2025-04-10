package models

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type Health struct {
	Status string `json:"status,omitempty"`
	Time   string `json:"time,omitempty"`
}
