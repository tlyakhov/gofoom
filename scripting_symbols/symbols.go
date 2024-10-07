// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package scripting_symbols

//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/archetypes
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/concepts
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/constants
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/containers
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/controllers
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/dynamic
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/ecs

//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/components/behaviors
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/components/core
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/components/materials
//go:generate $GOPATH/bin/yaegi extract --name scripting_symbols tlyakhov/gofoom/components/selection

import (
	"reflect"
	"tlyakhov/gofoom/ecs"
)

// Symbols variable stores the map of script symbols per package.
var Symbols = map[string]map[string]reflect.Value{}

// MapTypes variable contains a map of functions which have an interface{} as parameter but
// do something special if the parameter implements a given interface.
var MapTypes = map[reflect.Value][]reflect.Type{}

// See core.Script for usage
func init() {
	Symbols["github.com/tlyakhov/gofoom/scripting_symbols/scripting_symbols"] = map[string]reflect.Value{
		"Symbols": reflect.ValueOf(Symbols),
	}
	Symbols["."] = map[string]reflect.Value{
		"MapTypes": reflect.ValueOf(MapTypes),
	}
	// We have to set this here to avoid a circular reference.
	ecs.Types().InterpSymbols = Symbols
}
