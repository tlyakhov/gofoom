// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ui

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

type Config struct {
	LabelColor    concepts.Vector4
	SelectedColor concepts.Vector4
	BGColor       concepts.Vector4
	WidgetColor   concepts.Vector4
	ShadowColor   concepts.Vector4
	TextStyle     *render.TextStyle
	Padding       int
	ShadowText    bool
	OnChanged     func(IWidget)
}
type UI struct {
	Config
	*render.Renderer

	Page *Page

	cols, rows int
	textBuffer []textElement
}

func (ui *UI) SetPage(page *Page) {
	if ui.Page != nil {
		for _, item := range ui.Page.Widgets {
			w := item.GetWidget()
			w.highlight.Detach(ecs.Simulation)
		}
		ui.Page.tooltipAlpha.Detach(ecs.Simulation)
	}

	ui.Page = page

	if page == nil {
		return
	}

	page.tooltipAlpha.Attach(ecs.Simulation)
	a := page.tooltipAlpha.NewAnimation()
	page.tooltipAlpha.Animation = a
	a.Duration = 50
	a.Start = 0
	a.End = 1.0
	a.TweeningFunc = dynamic.EaseInOut2
	a.Lifetime = dynamic.AnimationLifetimeOnce
	a.Reverse = false
	a.Active = false
	a.Coordinates = dynamic.AnimationCoordinatesAbsolute

	for _, item := range page.Widgets {
		w := item.GetWidget()
		w.highlight.Attach(ecs.Simulation)
		w.highlight.SetAll(ui.WidgetColor)
		a := w.highlight.NewAnimation()
		w.highlight.Animation = a
		a.Duration = 250
		a.Start = w.highlight.Spawn
		a.End = ui.SelectedColor
		a.TweeningFunc = dynamic.EaseInOut2
		a.Lifetime = dynamic.AnimationLifetimeBounce
		a.Reverse = false
		a.Active = false
		a.Coordinates = dynamic.AnimationCoordinatesAbsolute
	}
}

func (ui *UI) Initialize() {
	ui.Padding = 2
	ui.LabelColor = concepts.Vector4{1, 1, 1, 1}
	ui.BGColor = concepts.Vector4{0.3, 0.3, 0.3, 1}
	ui.BGColor.MulSelf(0.5)
	ui.WidgetColor = concepts.Vector4{0.5, 0.5, 0.5, 1}
	ui.ShadowColor = concepts.Vector4{0.1, 0.1, 0.1, 1}
	ui.ShadowColor.MulSelf(0.5)
	ui.SelectedColor = concepts.Vector4{0, 0.431, 1, 1}
	ui.ShadowText = true

	ui.TextStyle = ui.NewTextStyle()
	ui.cols = ui.ScreenWidth / ui.TextStyle.CharWidth
	ui.rows = ui.ScreenHeight / ui.TextStyle.CharHeight

	if len(ui.textBuffer) != ui.cols*ui.rows {
		ui.textBuffer = make([]textElement, ui.cols*ui.rows)
	}
}

func (ui *UI) MoveUp() {
	if ui.Page == nil {
		return
	}
	ui.Page.SelectedItem--
	if ui.Page.SelectedItem < 0 {
		ui.Page.SelectedItem = 0
	}
	if w := ui.Page.SelectedWidget(); w.GetWidget().Tooltip != "" {
		ui.Page.tooltipQueue.PushBack(w)
	}

	for ui.Page.SelectedItem < ui.Page.ScrollPos {
		ui.Page.ScrollPos--
	}
}
func (ui *UI) MoveDown() {
	if ui.Page == nil {
		return
	}
	ui.Page.SelectedItem++
	if ui.Page.SelectedItem >= len(ui.Page.Widgets) {
		ui.Page.SelectedItem = len(ui.Page.Widgets) - 1
	}

	if w := ui.Page.SelectedWidget(); w.GetWidget().Tooltip != "" {
		ui.Page.tooltipQueue.PushBack(w)
	}

	for ui.Page.SelectedItem >= ui.Page.ScrollPos+ui.Page.VisibleWidgets {
		ui.Page.ScrollPos++
	}

}
func (ui *UI) Action() {
	if ui.Page == nil {
		return
	}
	item := ui.Page.Widgets[ui.Page.SelectedItem]
	switch w := item.(type) {
	case *Button:
		if w.Clicked != nil {
			w.Clicked(w)
		}
	case *Checkbox:
		w.Value = !w.Value
		if w.Checked != nil {
			w.Checked(w)
		}
		ui.OnChanged(item)
	}
}

func (ui *UI) EditLeft() {
	if ui.Page == nil {
		return
	}
	item := ui.Page.Widgets[ui.Page.SelectedItem]
	switch w := item.(type) {
	case *Slider:
		w.Value -= w.Step
		if w.Value < w.Min {
			w.Value = w.Min
		}
		if w.Moved != nil {
			w.Moved(w)
		}
		ui.OnChanged(item)
	}
}

func (ui *UI) EditRight() {
	if ui.Page == nil {
		return
	}
	item := ui.Page.Widgets[ui.Page.SelectedItem]
	switch w := item.(type) {
	case *Slider:
		w.Value += w.Step
		if w.Value > w.Max {
			w.Value = w.Max
		}
		if w.Moved != nil {
			w.Moved(w)
		}
		ui.OnChanged(item)
	}
}
