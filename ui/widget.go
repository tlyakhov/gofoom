// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"encoding/json"
	"fmt"
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
	return result
}

func (p *Page) Serialize() map[string]any {
	result := map[string]any{
		"Title": p.Title,
	}
	jsonWidgets := make(map[string]any)
	for _, w := range p.Widgets {
		switch ww := w.(type) {
		case *Slider:
			jsonWidgets[ww.ID] = ww.Serialize()
		case *Checkbox:
			jsonWidgets[ww.ID] = ww.Serialize()
		}
	}
	if len(jsonWidgets) == 0 {
		return nil
	}
	result["Widgets"] = jsonWidgets
	return result
}

func (p *Page) Construct(data map[string]any) {
	vw := data["Widgets"]
	if vw == nil {
		return
	}
	if jsonWidgets, ok := vw.(map[string]any); ok {
		for _, w := range p.Widgets {
			if v, ok := jsonWidgets[w.GetWidget().ID]; ok {
				if jsonWidget, ok := v.(map[string]any); ok {
					switch ww := w.(type) {
					case *Slider:
						ww.Construct(jsonWidget)
					case *Checkbox:
						ww.Construct(jsonWidget)
					}
				}
			}
		}
	}
}

func SaveSettings(filename string, pages ...*Page) {
	jsonData := make(map[string]any, 0)

	for _, page := range pages {
		jsonPage := page.Serialize()
		if jsonPage == nil {
			continue
		}
		jsonData[page.Title] = jsonPage
	}

	bytes, err := json.MarshalIndent(jsonData, "", "  ")

	if err != nil {
		panic(err)
	}

	os.WriteFile(filename, bytes, os.ModePerm)
}

func LoadSettings(filename string, pages ...*Page) error {
	fileContents, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	var parsed any
	err = json.Unmarshal(fileContents, &parsed)
	if err != nil {
		return err
	}

	var jsonPages map[string]any
	var ok bool
	if jsonPages, ok = parsed.(map[string]any); !ok || jsonPages == nil {
		return fmt.Errorf("JSON settings root must be an object")
	}

	for _, page := range pages {
		if v, ok := jsonPages[page.Title]; ok {
			if jsonPage, ok := v.(map[string]any); ok {
				page.Construct(jsonPage)
			}
		}
	}
	return nil
}
