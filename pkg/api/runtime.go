package api

type Object interface {
	GetID() string
}

type DataObject struct {
	Data Object `json:"data"`
}

func WrapObject(obj Object) *DataObject {
	return &DataObject{Data: obj}
}