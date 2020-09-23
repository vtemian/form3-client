package api

import (
	"fmt"
	"reflect"
)

type Scheme struct {
	objToType    map[string]reflect.Type
	typeToObj    map[reflect.Type]string
	objEndpoints map[string]string
}

func NewScheme() *Scheme {
	return &Scheme{
		objToType:    map[string]reflect.Type{},
		typeToObj:    map[reflect.Type]string{},
		objEndpoints: map[string]string{},
	}
}

func (s *Scheme) TypeName(obj Object) string {
	typeObj := realTypeOf(obj)
	return typeObj.String()
}

func realTypeOf(obj interface{}) reflect.Type {
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		return reflect.Indirect(reflect.ValueOf(obj)).Type()
	}

	return reflect.TypeOf(obj)
}

func (s *Scheme) Register(obj Object, endpoint string) {
	typeObj := reflect.TypeOf(obj)
	typeName := typeObj.String()

	s.typeToObj[typeObj] = typeName
	s.objToType[typeName] = typeObj
	s.objEndpoints[typeName] = endpoint
}

var missingObjTypeFmt = "missing type %s from scheme"

func (s *Scheme) NewObj(kind string) (Object, error) {
	reflectType, exists := s.objToType[kind]

	if !exists {
		return nil, fmt.Errorf(missingObjTypeFmt, kind)
	}

	obj, ok := (reflect.New(reflectType).Interface()).(Object)
	if !ok {
		return nil, fmt.Errorf("%s doesn't implement interface Object", reflectType)
	}

	return obj, nil
}

func (s *Scheme) NewDataObj(kind string) (*DataObject, error) {
	obj, err := s.NewObj(kind)
	if err != nil {
		return nil, err
	}

	result := &DataObject{Data: obj}
	return result, nil
}

var Schema = NewScheme()

func init() {
	Schema.Register(Account{}, "organisation/accounts/%s")
	Schema.Register(AccountList{}, "organisation/accounts")
}
