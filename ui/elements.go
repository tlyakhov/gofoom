// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"encoding/json"
	"os"
	"tlyakhov/gofoom/concepts"
)

type textElement struct {
	Rune    rune
	Color   *concepts.Vector4
	BGColor *concepts.Vector4
	Shadow  bool
}

type IWidget interface {
	GetWidget() *Widget
}

type Widget struct {
	ID        string
	Label     string
	Tooltip   string
	highlight concepts.DynamicValue[concepts.Vector4]
}

type Page struct {
	IsDialog     bool
	Title        string
	Widgets      []IWidget
	SelectedItem int
	Parent       *Page
}

type Button struct {
	Widget

	Clicked func(b *Button)
}

func (e *Widget) GetWidget() *Widget {
	return e
}

func (w *Widget) Serialize() map[string]any {
	result := make(map[string]any)
	result["ID"] = w.ID
	return result
}

func (p *Page) Serialize() map[string]any {
	result := map[string]any{
		"Title": p.Title,
	}
	jsonWidgets := make([]map[string]any, 0, len(p.Widgets))
	for _, w := range p.Widgets {
		switch ww := w.(type) {
		case *Slider:
			jsonWidgets = append(jsonWidgets, ww.Serialize())
		case *Checkbox:
			jsonWidgets = append(jsonWidgets, ww.Serialize())
		}
	}
	if len(jsonWidgets) == 0 {
		return nil
	}
	result["Widgets"] = jsonWidgets
	return result
}

func SaveSettings(filename string, pages ...*Page) {
	jsonData := make([]any, 0)

	for _, page := range pages {
		jsonPage := page.Serialize()
		if jsonPage == nil {
			continue
		}
		jsonData = append(jsonData, jsonPage)
	}

	bytes, err := json.MarshalIndent(jsonData, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}
