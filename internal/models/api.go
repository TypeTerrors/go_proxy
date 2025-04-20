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

type AddNewProxy struct {
	From string `json:"from"`
	To   string `json:"to"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
type PatchOldProxy struct {
	From string `json:"from"`
	To   string `json:"to"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
type DelOldProxy struct {
	From string `json:"from"`
}
type RedirectionRecords struct {
	From string `json:"from"`
	To   string `json:"to"`
}
