// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"image"
	"strings"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/gogpu/gg"
	"github.com/spf13/cast"
	"github.com/gogpu/gg/text"
	"golang.org/x/image/font/gofont/gomono"
)

type Text struct {
	ecs.Attached `editable:"^"`
	Rendered     *Image

	Label       string  `editable:"Label" edit_type:"multi-line-string"`
	LineSpacing float64 `editable:"Line Spacing"`

	Color dynamic.DynamicValue[concepts.Vector4] `editable:"Color"`
}

func (t *Text) Shareable() bool { return true }

func (t *Text) OnDelete() {
	defer t.Attached.OnDelete()
	if t.IsAttached() {
		t.Color.Detach(ecs.Simulation)
	}
}
func (t *Text) OnAttach() {
	t.Attached.OnAttach()
	t.Color.Attach(ecs.Simulation)
}

func (t *Text) RasterizeText() {
	padding := 4.0
	fontSource, _ := text.NewFontSource(gomono.TTF)
	face := fontSource.Face(16.0)
	// For measuring text
	c := gg.NewContext(0, 0)
	c.SetFont(face)
	w, h := c.MeasureMultilineString(t.Label, t.LineSpacing)
	w += padding * 2
	h += padding * 2
	// Actual image & context
	c = gg.NewContext(int(w), int(h))
	c.SetFont(face)
	// Color gets applied at render time
	c.SetRGBA(1, 1, 1, 1)
	split := strings.Split(t.Label, "\n")
	// height formula is from MeasureMultilineString
	lineHeight := c.Font().Metrics().LineHeight() * t.LineSpacing
	th := float64(len(split)) * lineHeight
	th -= (t.LineSpacing - 1) * c.Font().Metrics().LineHeight()
	for i, line := range split {
		c.DrawStringAnchored(line, w*0.5, padding+(h-th)*0.5+float64(i)*lineHeight, 0.5, 0.5)
	}
	img := new(Image)
	img.Width = uint32(w)
	img.Height = uint32(h)
	img.PixelsRGBA = make([]uint32, int(img.Width)*int(img.Height))
	img.PixelsLinear = make([]concepts.Vector4, int(img.Width)*int(img.Height))
	img.GenerateMipMaps = false
	img.Filter = true
	rgba := c.Image().(*image.RGBA)
	for i := 0; i < len(rgba.Pix)/4; i++ {
		// Only need the alpha, the rest is for editing/debugging
		a := uint32(rgba.Pix[i*4+3])
		r := uint32(a)
		img.PixelsRGBA[i] = ((r & 0xFF) << 24) | ((r & 0xFF) << 16) | ((r & 0xFF) << 8) | (a & 0xFF)
		img.PixelsLinear[i][3] = float64(a) / 255
	}

	t.Rendered = img
	t.Rendered.Image = rgba
}

func (t *Text) Sample(x, y float64, sw, sh uint32, result *concepts.Vector4) {
	if t.Rendered == nil {
		t.RasterizeText()
	}
	t.Rendered.Sample(x, y, sw, sh, result)
	result[0] = t.Color.Render[0] * result[3]
	result[1] = t.Color.Render[1] * result[3]
	result[2] = t.Color.Render[2] * result[3]
	result[3] *= t.Color.Render[3]
}

var defaultTextColor = map[string]any{"Spawn": "1,1,1,1"}

func (t *Text) Construct(data map[string]any) {
	t.Attached.Construct(data)
	t.Color.Construct(defaultTextColor)
	t.LineSpacing = 1.05

	if data == nil {
		return
	}

	if v, ok := data["LineSpacing"]; ok {
		t.LineSpacing = cast.ToFloat64(v)
	}

	if v, ok := data["Label"]; ok {
		t.Label = v.(string)
		t.RasterizeText()
	}

	if v, ok := data["Color"]; ok {
		t.Color.Construct(v)
	}
}

func (t *Text) Serialize() map[string]any {
	result := t.Attached.Serialize()
	if t.Label != "" {
		result["Label"] = t.Label
	}
	result["Color"] = t.Color.Serialize()
	result["LineSpacing"] = t.LineSpacing
	return result
}
