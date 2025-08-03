// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"text/template"
	"tlyakhov/gofoom/ecs"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type ScriptParam struct {
	Name     string
	TypeName string
}

type Script struct {
	Code string `editable:"Code" edit_type:"multi-line-string"`

	ErrorMessage string
	interp       *interp.Interpreter
	Vars         map[string]any
	Params       []ScriptParam
	runFunc      any
	execCode     string
}

var scriptTemplate *template.Template

func init() {
	var err error
	scriptTemplate, err = template.New("script").Parse(`
	package main

	import "tlyakhov/gofoom/archetypes"
	import "tlyakhov/gofoom/components/behaviors"
	import "tlyakhov/gofoom/components/character"
	import "tlyakhov/gofoom/components/core"
	import "tlyakhov/gofoom/components/inventory"
	import "tlyakhov/gofoom/components/materials"
	import "tlyakhov/gofoom/concepts"
	import "tlyakhov/gofoom/constants"
	import "tlyakhov/gofoom/containers"
	import "tlyakhov/gofoom/controllers"
	import "tlyakhov/gofoom/ecs"
	import "log"
	import "fmt"

	func Do(s *core.Script) {
		{{range .Params}}
			var {{.Name}} {{.TypeName}}
			if s.Vars["{{.Name}}"] != nil {
				{{.Name}} = s.Vars["{{.Name}}"].({{.TypeName}})
			}
		{{end}}

		{{.Code}}
	}

	`)
	if err != nil {
		log.Printf("core.Script: error building script template: %v", err)
	}
}

func (s *Script) Compile() {
	/*if !s.IsAttached() {
		log.Println("Script.Compile: Universe is nil. Stack trace:")
		log.Println(concepts.StackTrace())
		return
	}*/

	s.ErrorMessage = ""
	s.interp = interp.New(interp.Options{})
	s.interp.Use(stdlib.Symbols)
	s.interp.Use(ecs.Types().InterpSymbols)

	var buf bytes.Buffer
	err := scriptTemplate.Execute(&buf, s)
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error building script template %v: %v", s.Code, err)
		s.interp = nil
		log.Printf("%v", s.ErrorMessage)
		return
	}
	s.execCode = buf.String()

	_, err = s.interp.Eval(s.execCode)
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.execCode, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}
	// log.Printf("%v", codeTemplate)
	f, err := s.interp.Eval("main.Do")
	if err != nil {
		s.ErrorMessage += fmt.Sprintf("Error compiling script %v: %v", s.execCode, err)
		log.Printf("%v", s.ErrorMessage)
		return
	}

	s.runFunc = f.Interface()
}

func (s *Script) IsCompiled() bool {
	return s.interp != nil && s.runFunc != nil
}

func (s *Script) IsEmpty() bool {
	return len(s.Code) == 0
}

func (s *Script) OnAttach() {
}

func (s *Script) IsAttached() bool {
	return true
}

func (s *Script) Construct(data map[string]any) {
	s.ErrorMessage = ""
	s.Vars = maps.Clone(ecs.Types().ExprEnv)

	if data == nil {
		return
	}

	if v, ok := data["Params"]; ok {
		s.Params = v.([]ScriptParam)
	}

	if v, ok := data["Code"]; ok {
		s.Code = v.(string)
	}
}

func (s *Script) Serialize() map[string]any {
	data := make(map[string]any)
	data["Code"] = s.Code
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
	if !s.IsCompiled() {
		return false
	}
	// This handles panics inside the interpreter. It's a bit aggressive, but
	// saves the game from crashing due to bad scripting
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		log.Printf("Script error: %v", recovered)
		log.Printf("in code: %v", s.execCode)
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
