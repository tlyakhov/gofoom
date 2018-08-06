package render

import (
	"image/color"
	"io/ioutil"
	"os"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type Font struct {
	font.Face
	atlas *text.Atlas
}

func NewFont(path string, size float64) (*Font, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		return nil, err
	}

	f := &Font{Face: truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})}
	f.atlas = text.NewAtlas(f.Face, text.ASCII)
	return f, nil
}

func (f *Font) Draw(win *pixelgl.Window, x, y float64, c color.Color, s string) {
	txt := text.New(pixel.V(x, y), f.atlas)
	txt.Color = c
	txt.WriteString(s)
	txt.Draw(win, pixel.IM.Moved(pixel.V(x, y)))
}
