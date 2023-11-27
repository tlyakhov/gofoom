package core

import (
	"log"
	"maps"
	"reflect"
	"strings"
	"tlyakhov/gofoom/concepts"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expression struct {
	Code     string `editable:"Code"`
	compiled *vm.Program
	env      map[string]any
}

func (e *Expression) setterWithList(components []concepts.Attachable, path string, v any) bool {
	if components == nil {
		return false
	}
	sPath := strings.Split(path, ".")
	if len(sPath) < 2 {
		log.Printf("Expression::setterWithList error: path should have at least a component and field")
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
				log.Printf("Expression::setterWithList error: path %v had an invalid field %v", path, sPath[i])
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
		log.Printf("Error compiling expression: %v", err)
		return
	}
	e.compiled = c
}

func (f *Expression) Serialize() string {
	return f.Code
}

func (f *Expression) Valid(ref *concepts.EntityRef) bool {
	if f.compiled == nil {
		return false
	}
	f.env["Ref"] = ref
	result, err := expr.Run(f.compiled, f.env)
	if err != nil {
		log.Printf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}

func (f *Expression) Act(ref *concepts.EntityRef) bool {
	if f.compiled == nil {
		return false
	}
	f.env["Ref"] = ref
	result, err := expr.Run(f.compiled, f.env)
	if err != nil {
		log.Printf("Expression error: %v", err)
		return false
	}
	return result.(bool)
}
