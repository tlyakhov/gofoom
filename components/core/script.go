// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"fmt"
	"log"
	"maps"
	"strings"
	"tlyakhov/gofoom/ecs"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type ScriptParam struct {
	Name     string
	TypeName string
}

type Script struct {
	Code  string      `editable:"Code" edit_type:"multi-line-string"`
	Style ScriptStyle `editable:"Style"`

	ErrorMessage string
	interp       *interp.Interpreter
	Vars         map[string]any
	Params       []ScriptParam
	ECS          *ecs.ECS
	System       bool
	runFunc      any
}

// TODO: Refactor this to use text/Template
func (s *Script) codeParams(sb *strings.Builder) {
	for _, p := range s.Params {
		sb.WriteString("var ")
		sb.WriteString(p.Name)
		sb.WriteString(" ")
		sb.WriteString(p.TypeName)
		sb.WriteString("\n")
		sb.WriteString(`if s.Vars["`)
		sb.WriteString(p.Name)
		sb.WriteString("\"] != nil {\n")
		sb.WriteString(p.Name)
		sb.WriteString(` = s.Vars["`)
		sb.WriteString(p.Name)
		sb.WriteString(`"].(`)
		sb.WriteString(p.TypeName)
		sb.WriteString(")\n}\n")
	}
}

func (s *Script) codeCommonHeader(sb *strings.Builder) {
	sb.WriteString(`package main
	import "tlyakhov/gofoom/archetypes"
	import "tlyakhov/gofoom/components/behaviors"
	import "tlyakhov/gofoom/components/core"
	import "tlyakhov/gofoom/components/materials"
	import "tlyakhov/gofoom/concepts"
	import "tlyakhov/gofoom/containers"
	import "tlyakhov/gofoom/constants"
	import "tlyakhov/gofoom/ecs"
	import "log"
	import "fmt"
	`)
}
func (s *Script) codeBoolExpr(sb *strings.Builder) {
	s.codeCommonHeader(sb)
	sb.WriteString(`
func Do(s *core.Script) bool {
	`)
	s.codeParams(sb)
	sb.WriteString(`
return (` + s.Code + `)
	}`)
}

func (s *Script) codeStatement(sb *strings.Builder) {
	s.codeCommonHeader(sb)
	sb.WriteString(`
func Do(s *core.Script) {
	`)
	s.codeParams(sb)
	sb.WriteString(`
	` + s.Code + `
	}`)
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
		sb := new(strings.Builder)
		s.codeBoolExpr(sb)
		codeTemplate = sb.String()
	case ScriptStyleStatement:
		sb := new(strings.Builder)
		s.codeStatement(sb)
		codeTemplate = sb.String()
	}

	_, err := s.interp.Eval(codeTemplate)
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.Code, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}
	//log.Printf("%v", codeTemplate)
	f, err := s.interp.Eval("main.Do")
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.Code, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}

	s.runFunc = f.Interface()
}

func (s *Script) AttachECS(db *ecs.ECS) {
	s.ECS = db
}

func (s *Script) GetECS() *ecs.ECS {
	return s.ECS
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

	// !!! Should happen prior to .Compile()
	if v, ok := data["Params"]; ok {
		s.Params = v.([]ScriptParam)
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
