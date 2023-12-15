package core

import (
	"fmt"
	"log"
	"maps"
	"tlyakhov/gofoom/concepts"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Script struct {
	Code  string      `editable:"Code"`
	Style ScriptStyle `editable:"Style"`

	ErrorMessage string
	interp       *interp.Interpreter
	Vars         map[string]any
	DB           *concepts.EntityComponentDB
	runFunc      any
}

func (e *Script) codeCommonHeader() string {
	return `package main
	import "tlyakhov/gofoom/components/behaviors"
	import "tlyakhov/gofoom/components/core"
	import "tlyakhov/gofoom/components/materials"
	import "tlyakhov/gofoom/components/sectors"
	import "tlyakhov/gofoom/concepts"
	import "tlyakhov/gofoom/constants"
	`
}
func (e *Script) codeBoolExpr() string {
	return e.codeCommonHeader() + `
func Do(s *core.Script) bool {
	return (` + e.Code + `)
}`
}

func (e *Script) codeStatement() string {
	return e.codeCommonHeader() + `
func Do(s *core.Script) {
	` + e.Code + `
}`
}

func (e *Script) Compile() {
	e.ErrorMessage = ""
	e.interp = interp.New(interp.Options{})
	e.interp.Use(stdlib.Symbols)
	e.interp.Use(concepts.DbTypes().InterpSymbols)
	var codeTemplate string
	switch e.Style {
	case ScriptStyleRaw:
		codeTemplate = e.Code
	case ScriptStyleBoolExpr:
		codeTemplate = e.codeBoolExpr()
	case ScriptStyleStatement:
		codeTemplate = e.codeStatement()
	}
	_, err := e.interp.Eval(codeTemplate)
	if err != nil {
		e.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", e.Code, err)
		log.Printf("%v", e.ErrorMessage)
		return
	}
	f, err := e.interp.Eval("main.Do")
	if err != nil {
		e.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", e.Code, err)
		log.Printf("%v", e.ErrorMessage)
		return
	}

	e.runFunc = f.Interface()
}

func (s *Script) Construct(db *concepts.EntityComponentDB, data map[string]any) {
	s.DB = db
	s.ErrorMessage = ""
	s.Vars = maps.Clone(concepts.DbTypes().ExprEnv)

	if data == nil {
		return
	}

	if v, ok := data["Style"]; ok {
		s.Style, _ = ScriptStyleString(v.(string))
	}

	if v, ok := data["Code"]; ok {
		s.Code = v.(string)
		s.Compile()
	}
}

func (s *Script) Serialize() map[string]any {
	data := make(map[string]any)
	data["Code"] = s.Code
	data["Style"] = s.Style
	return data
}

func (s *Script) Ref(name string) *concepts.EntityRef {
	if s.Vars[name] == nil {
		return nil
	}
	if ref, ok := s.Vars[name].(*concepts.EntityRef); ok {
		return ref
	}
	return nil
}

func (s *Script) Valid() bool {
	if s.interp == nil || s.runFunc == nil {
		return false
	}
	if f, ok := s.runFunc.(func(*Script) bool); ok {
		return f(s)
	} else {
		s.ErrorMessage = "Error running script, 'Run' function has the wrong signature."
		return false
	}
}

func (s *Script) Act() bool {
	if s.interp == nil || s.runFunc == nil {
		return false
	}
	if f, ok := s.runFunc.(func(*Script)); ok {
		f(s)
		return true
	} else {
		s.ErrorMessage = "Error running script, 'Run' function has the wrong signature."
		return false
	}
}
