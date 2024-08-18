// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"tlyakhov/gofoom/ecs"

	"github.com/gammazero/deque"
)

type Page struct {
	IsDialog     bool
	Title        string
	Widgets      []IWidget
	SelectedItem int
	Parent       *Page

	ScrollPos      int
	VisibleWidgets int

	tooltipCurrent IWidget
	tooltipAlpha   ecs.DynamicValue[float64]
	tooltipQueue   deque.Deque[IWidget]
}

func (p *Page) SelectedWidget() IWidget {
	return p.Widgets[p.SelectedItem]
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
