package api

type Object interface {
	GetID() string
	GetVersion() int
}

type Links struct {
	First string `json:"first"`
	Next  string `json:"next"`
	Last  string `json:"last"`
	Self  string `json:"self"`
}

type DataObject struct {
	Data  Object `json:"data"`
	Links Links  `json:"links"`
}

func WrapObject(obj Object) *DataObject {
	return &DataObject{Data: obj, Links: Links{}}
}
