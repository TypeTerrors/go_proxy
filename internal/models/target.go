package models

type AddNewProxy struct {
	From string `json:"from"`
	To   string `json:"to"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}
type RedirectionRecords struct {
	From string `json:"from"`
	To   string `json:"to"`
}
