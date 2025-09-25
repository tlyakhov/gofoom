// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"tlyakhov/gofoom/dynamic"

	"github.com/gammazero/deque"
)

type Page struct {
	IsDialog     bool
	Title        string
	Widgets      []IWidget
	SelectedItem int
	Parent       *Page
	Apply        func(p *Page)

	ScrollPos       int
	VisibleWidgets  int
	HijackAllInputs bool // For input binding

	mapped         map[string]IWidget
	tooltipCurrent IWidget
	tooltipAlpha   dynamic.DynamicValue[float64]
	tooltipQueue   deque.Deque[IWidget]
	lastMeasuredX  int
	lastMeasuredY  int
}

func (p *Page) Initialize() {
	p.mapped = make(map[string]IWidget)

	for _, w := range p.Widgets {
		p.mapped[w.GetWidget().ID] = w
		w.GetWidget().page = p
	}
}

func (p *Page) Widget(id string) IWidget {
	return p.mapped[id]
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
		case *InputBinding:
			jsonWidgets[ww.ID] = ww.Serialize()
		}
	}
	if len(jsonWidgets) == 0 {
		return nil
	}
	result["Widgets"] = jsonWidgets
	return result
}

func (p *Page) constructWidget(w IWidget, jsonWidget map[string]any) {
	switch ww := w.(type) {
	case *Slider:
		ww.Construct(jsonWidget)
		if ww.Moved != nil {
			ww.Moved(ww)
		}
	case *Checkbox:
		ww.Construct(jsonWidget)
		if ww.Checked != nil {
			ww.Checked(ww)
		}
	case *InputBinding:
		ww.Construct(jsonWidget)
		if ww.Changed != nil {
			ww.Changed(ww)
		}
	}
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
					p.constructWidget(w, jsonWidget)
				}
			}
		}
	}
	if p.Apply != nil {
		p.Apply(p)
	}
}
