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
	env          map[string]any
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
	ref := e.env["Ref"].(*concepts.EntityRef)
	if ref.Nil() {
		return false
	}
	return e.setterWithList(ref.DB.EntityComponents[entity], path, v)
}
func (e *Expression) setter(path string, v any) bool {
	ref := e.env["Ref"].(*concepts.EntityRef)
	if ref.Nil() {
		return false
	}
	return e.setterWithList(ref.All(), path, v)
}

func (e *Expression) Construct(code string) {
	e.env = maps.Clone(concepts.DbTypes().ExprEnv)
	e.env["Ref"] = (*concepts.EntityRef)(nil)
	e.env["Set"] = e.setter
	e.env["EntitySet"] = e.setterWithEntity
	e.Code = code
	c, err := expr.Compile(code, expr.Env(e.env), expr.AsBool())
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Error compiling expression: %v", err)
		return
	}
	e.compiled = c
}

func (e *Expression) Serialize() string {
	return e.Code
}

func (e *Expression) Valid(ref *concepts.EntityRef) bool {
	if e.compiled == nil {
		return false
	}
	e.env["Ref"] = ref
	result, err := expr.Run(e.compiled, e.env)
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}

func (e *Expression) Act(ref *concepts.EntityRef) bool {
	if e.compiled == nil {
		return false
	}
	e.env["Ref"] = ref
	result, err := expr.Run(e.compiled, e.env)
	if err != nil {
		e.ErrorMessage = fmt.Sprintf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}
