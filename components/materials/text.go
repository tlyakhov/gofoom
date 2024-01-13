package materials

import (
	"image"
	"strings"
	"tlyakhov/gofoom/concepts"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/inconsolata"
)

type Text struct {
	concepts.Attached `editable:"^"`
	Rendered          *Image

	Label       string  `editable:"Label" edit_type:"multi-line-string"`
	LineSpacing float64 `editable:"Line Spacing"`

	Color concepts.SimVariable[concepts.Vector4] `editable:"Color"`
}

var TextComponentIndex int

func init() {
	TextComponentIndex = concepts.DbTypes().Register(Text{}, TextFromDb)
}

func TextFromDb(entity *concepts.EntityRef) *Text {
	if asserted, ok := entity.Component(TextComponentIndex).(*Text); ok {
		return asserted
	}
	return nil
}

func (t *Text) SetDB(db *concepts.EntityComponentDB) {
	if t.DB != nil {
		t.Color.Detach(t.DB.Simulation)
	}
	t.Attached.SetDB(db)
	t.Color.Attach(db.Simulation)
}

func (t *Text) RasterizeText() {
	padding := 4.0
	// For measuring text
	c := gg.NewContext(1, 1)
	c.SetFontFace(inconsolata.Regular8x16)
	w, h := c.MeasureMultilineString(t.Label, t.LineSpacing)
	w += padding * 2
	h += padding * 2
	// Actual image & context
	c = gg.NewContext(int(w), int(h))
	c.SetFontFace(inconsolata.Regular8x16)
	// Color gets applied at render time
	c.SetRGBA255(255, 255, 255, 255)
	split := strings.Split(t.Label, "\n")
	// height formula is from MeasureMultilineString
	lineHeight := c.FontHeight() * t.LineSpacing
	th := float64(len(split)) * lineHeight
	th -= (t.LineSpacing - 1) * c.FontHeight()
	for i, line := range split {
		c.DrawStringAnchored(line, w*0.5, padding+(h-th)*0.5+float64(i)*lineHeight, 0.5, 0.5)
	}
	img := new(Image)
	img.Width = uint32(w)
	img.Height = uint32(h)
	img.Data = make([]uint32, int(img.Width)*int(img.Height))
	img.GenerateMipMaps = false
	img.Filter = true
	rgba := c.Image().(*image.RGBA)
	for i := 0; i < len(rgba.Pix)/4; i++ {
		// Only need the alpha, the rest is for editing/debugging
		a := uint32(rgba.Pix[i*4+3])
		r := uint32(255 * a)
		img.Data[i] = ((r & 0xFF) << 24) | ((r & 0xFF) << 16) | ((r & 0xFF) << 8) | (a & 0xFF)
	}

	t.Rendered = img
	t.Rendered.Image = rgba
}

func (t *Text) Sample(x, y float64, scale float64) concepts.Vector4 {
	if t.Rendered == nil {
		t.RasterizeText()
	}
	c := t.Rendered.Sample(x, y, scale)
	c[0] = t.Color.Render[0] * c[3]
	c[1] = t.Color.Render[1] * c[3]
	c[2] = t.Color.Render[2] * c[3]
	c[3] *= t.Color.Render[3]
	return c
}

func (t *Text) Construct(data map[string]any) {
	t.Attached.Construct(data)
	t.Color.Set(concepts.Vector4{1, 1, 1, 1})
	t.LineSpacing = 1.05

	if data == nil {
		return
	}

	if v, ok := data["LineSpacing"]; ok {
		t.LineSpacing = v.(float64)
	}

	if v, ok := data["Label"]; ok {
		t.Label = v.(string)
		t.RasterizeText()
	}

	if v, ok := data["Color"]; ok {
		t.Color.Construct(v.(map[string]any))
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
