package triggers

import (
	"log"
	"maps"
	"tlyakhov/gofoom/concepts"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type Expression struct {
	Code     string
	compiled *vm.Program
	env      map[string]any
}

func (f *Expression) Construct(code string) {
	f.env = maps.Clone(concepts.DbTypes().ExprEnv)
	f.env["Ref"] = (*concepts.EntityRef)(nil)
	f.Code = code
	c, err := expr.Compile(code, expr.Env(f.env), expr.AsBool())
	if err != nil {
		log.Printf("Error compiling expression: %v", err)
		return
	}
	f.compiled = c
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
