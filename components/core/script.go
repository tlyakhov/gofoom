// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

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
	Code  string      `editable:"Code" edit_type:"multi-line-string"`
	Style ScriptStyle `editable:"Style"`

	ErrorMessage string
	interp       *interp.Interpreter
	Vars         map[string]any
	DB           *concepts.EntityComponentDB
	runFunc      any
}

func (e *Script) codeCommonHeader() string {
	return `package main
	import "tlyakhov/gofoom/archetypes"
	import "tlyakhov/gofoom/components/behaviors"
	import "tlyakhov/gofoom/components/core"
	import "tlyakhov/gofoom/components/materials"
	import "tlyakhov/gofoom/concepts"
	import "tlyakhov/gofoom/constants"
	import "log"
	import "fmt"
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

func (s *Script) SetDB(db *concepts.EntityComponentDB) {
	s.DB = db
}

func (s *Script) GetDB() *concepts.EntityComponentDB {
	return s.DB
}

func (s *Script) Construct(data map[string]any) {
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

func (s *Script) Entity(name string) concepts.Entity {
	if s.Vars[name] == nil {
		return 0
	}
	if entity, ok := s.Vars[name].(concepts.Entity); ok {
		return entity
	}
	return 0
}

func (s *Script) Act() bool {
	if s.interp == nil || s.runFunc == nil {
		return false
	}
	// This handles panics inside the interpreter. It's a bit aggressive, but
	// saves the game from crashing due to bad scripting
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		s.ErrorMessage = fmt.Sprintf("Script error: %v", recovered)
		s.interp = nil
		s.runFunc = nil
	}()

	if f, ok := s.runFunc.(func(*Script)); ok {
		f(s)
		return true
	} else {
		s.ErrorMessage = "Error running script, 'Run' function has the wrong signature."
		return false
	}
}
