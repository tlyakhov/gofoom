package render

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type Font struct {
	font.Face
	atlas *text.Atlas
}

func NewFont(path string, size float64) (*Font, error) {
	f := &Font{Face: basicfont.Face7x13}
	f.atlas = text.NewAtlas(f.Face, text.ASCII)
	return f, nil

	/*
	   file, err := os.Open(path)

	   	if err != nil {
	   		return nil, err
	   	}

	   defer file.Close()

	   bytes, err := io.ReadAll(file)

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
	*/
}

func (f *Font) Draw(win *pixelgl.Window, x, y float64, c color.Color, s string) {
	// log.Printf("Font draw: %v\n", f.atlas.Glyph(text.ASCII[65]))
	txt := text.New(pixel.V(x, y), f.atlas)
	txt.Color = c

	txt.WriteString(s)
	txt.Draw(win, pixel.IM.Moved(pixel.V(x, y)).Scaled(pixel.Vec{}, 4))
}
