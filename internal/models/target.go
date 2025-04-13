package models

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
