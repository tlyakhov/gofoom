package core

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"tlyakhov/gofoom/concepts"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expression struct {
	Code         string `editable:"Code"`
	ErrorMessage string
	compiled     *vm.Program
	Env          map[string]any
	DB           *concepts.EntityComponentDB
}

func (e *Expression) setterWithList(components []concepts.Attachable, path string, v any) bool {
	if components == nil || e.ErrorMessage != "" {
		return false
	}
	sPath := strings.Split(path, ".")
	if len(sPath) < 2 {
		e.ErrorMessage = "Expression::setterWithList error: path should have at least a component and field"
		return false
	}
	if index, ok := concepts.DbTypes().IndexesNoPackage[sPath[0]]; ok {
		c := components[index]
		if c == nil {
			return false
		}
		trav := reflect.ValueOf(c).Elem()
		for i := 1; i < len(sPath); i++ {
			trav = trav.FieldByName(sPath[i])
			if !trav.IsValid() {
				e.ErrorMessage = fmt.Sprintf("Expression::setterWithList error: path %v had an invalid field %v", path, sPath[i])
				return false
			}
			if trav.Kind() == reflect.Ptr || trav.Kind() == reflect.Interface {
				trav = trav.Elem()
			}
		}
		trav.Set(reflect.ValueOf(v))
		return true
	}
	return false
}

func (e *Expression) setterWithEntity(entity uint64, path string, v any) bool {
	return e.setterWithList(e.DB.EntityComponents[entity], path, v)
}
func (e *Expression) setterWithRef(ref *concepts.EntityRef, path string, v any) bool {
	if ref.Nil() {
		return false
	}
	return e.setterWithList(ref.All(), path, v)
}

func (e *Expression) Construct(db *concepts.EntityComponentDB, code string) {
	e.DB = db
	e.ErrorMessage = ""
	e.Env = maps.Clone(concepts.DbTypes().ExprEnv)
	e.Env["Set"] = e.setterWithRef
	e.Env["SetEntity"] = e.setterWithEntity
	e.Code = code
	c, err := expr.Compile(code, expr.Env(e.Env), expr.AsBool())
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Error compiling expression: %v", err)
		return
	}
	e.compiled = c
}

func (e *Expression) Serialize() string {
	return e.Code
}

func (e *Expression) Valid() bool {
	if e.compiled == nil {
		return false
	}
	result, err := expr.Run(e.compiled, e.Env)
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}

func (e *Expression) Act() bool {
	if e.compiled == nil {
		return false
	}
	result, err := expr.Run(e.compiled, e.Env)
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}
