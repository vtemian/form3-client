package api

type Object interface {
	GetID() string
}

type DataObject struct {
	Data Object `json:"data"`
}
