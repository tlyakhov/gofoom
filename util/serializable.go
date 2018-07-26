package util

import (
	"reflect"

	"github.com/rs/xid"
)

type CommonFields struct {
	ID   string   `editable:"ID" edit_type:"string"`
	Tags []string `editable:"Tags" edit_type:"tags"`
}

func (cf *CommonFields) GenerateID(target interface{}) {
	cf.ID = reflect.ValueOf(target).Type().String() + "_" + xid.New().String()
}
