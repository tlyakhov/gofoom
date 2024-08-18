// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"fmt"
	"log"
	"maps"
	"tlyakhov/gofoom/ecs"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Script struct {
	Code  string      `editable:"Code" edit_type:"multi-line-string"`
	Style ScriptStyle `editable:"Style"`

	ErrorMessage string
	interp       *interp.Interpreter
	Vars         map[string]any
	DB           *ecs.ECS
	System       bool
	runFunc      any
}

func (s *Script) codeCommonHeader() string {
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
func (s *Script) codeBoolExpr() string {
	return s.codeCommonHeader() + `
func Do(s *core.Script) bool {
	return (` + s.Code + `)
}`
}

func (s *Script) codeStatement() string {
	return s.codeCommonHeader() + `
func Do(s *core.Script) {
	` + s.Code + `
}`
}

func (s *Script) Compile() {
	s.ErrorMessage = ""
	s.interp = interp.New(interp.Options{})
	s.interp.Use(stdlib.Symbols)
	s.interp.Use(ecs.Types().InterpSymbols)
	var codeTemplate string
	switch s.Style {
	case ScriptStyleRaw:
		codeTemplate = s.Code
	case ScriptStyleBoolExpr:
		codeTemplate = s.codeBoolExpr()
	case ScriptStyleStatement:
		codeTemplate = s.codeStatement()
	}

	_, err := s.interp.Eval(codeTemplate)
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.Code, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}
	f, err := s.interp.Eval("main.Do")
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.Code, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}

	s.runFunc = f.Interface()
}

func (s *Script) SetECS(db *ecs.ECS) {
	s.DB = db
}

func (s *Script) GetECS() *ecs.ECS {
	return s.DB
}

func (s *Script) IsSystem() bool {
	return s.System
}

func (s *Script) Construct(data map[string]any) {
	s.ErrorMessage = ""
	s.Vars = maps.Clone(ecs.Types().ExprEnv)

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

func (s *Script) Entity(name string) ecs.Entity {
	if s.Vars[name] == nil {
		return 0
	}
	if entity, ok := s.Vars[name].(ecs.Entity); ok {
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
