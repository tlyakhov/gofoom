// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"fmt"
	"image"
	"image/color"
	"os"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"tlyakhov/gofoom/concepts"

	"github.com/disintegration/gift"
	"github.com/fogleman/gg"
	"golang.org/x/image/font/inconsolata"
)

type mipMap struct {
	Width, Height uint32
	Data          []uint32
	Image         *image.RGBA
}

// Image represents an image that will be rendered in-game.
type Image struct {
	concepts.Attached `editable:"^"`

	Width, Height   uint32
	Source          string `editable:"Texture Source" edit_type:"file"`
	GenerateMipMaps bool   `editable:"Generate Mip Maps?" edit_type:"bool"`
	Filter          bool   `editable:"Filter?" edit_type:"bool"`
	Data            []uint32
	MipMaps         []mipMap
	Image           image.Image
}

var ImageComponentIndex int

func init() {
	ImageComponentIndex = concepts.DbTypes().Register(Image{}, ImageFromDb)
}

func ImageFromDb(db *concepts.EntityComponentDB, e concepts.Entity) *Image {
	if asserted, ok := db.Component(e, ImageComponentIndex).(*Image); ok {
		return asserted
	}
	return nil
}

func (img *Image) String() string {
	return "Image: " + img.Source
}

// Load a texture from a file (pre-processing mipmaps if set)
func (img *Image) Load() error {
	if img.Source == "" {
		return nil
	}

	// Load the image from a file...
	file, err := os.Open(img.Source)
	if err != nil {
		return err
	}
	defer file.Close()
	img.Image, _, err = image.Decode(file)
	if err != nil {
		return err
	}

	// Let's convert to 0-based NRGBA for speed/convenience.
	bounds := img.Image.Bounds()
	img.Width = uint32(bounds.Dx())
	img.Height = uint32(bounds.Dy())
	img.Data = make([]uint32, int(img.Width)*int(img.Height))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			index := uint32(x-bounds.Min.X) + uint32(y-bounds.Min.Y)*img.Width
			// Premultiplied alpha
			img.Data[index] = concepts.ColorToInt32PreMul(img.Image.At(x, y))
		}
	}
	img.generateMipMaps()
	//img.generateTestMipMaps()
	return nil
}

func (img *Image) generateTestMipMaps() {
	img.MipMaps = make([]mipMap, 0)
	w := img.Width
	h := img.Height
	for w > 4 && h > 4 {
		index := len(img.MipMaps)
		bg := (index + 1) * 255 / 6
		mm := mipMap{Width: w, Height: h, Data: make([]uint32, w*h)}
		face := inconsolata.Regular8x16
		c := gg.NewContext(int(w), int(h))
		c.SetFontFace(face)
		c.SetRGBA255(bg, bg, bg, 255)
		c.DrawRectangle(0, 0, float64(w), float64(h))
		c.Fill()
		c.SetRGBA255(255, 0, 0, 255)
		c.Translate(float64(w)*0.5, float64(h)*0.5)
		c.Scale(8.0/float64(index+1), 8.0/float64(index+1))
		c.DrawStringAnchored(fmt.Sprintf("Index: %v, w:%v,h:%v", index, w, h), 0, 0, 0.5, 0.5)
		rgba := c.Image().(*image.RGBA)
		for i := 0; i < len(rgba.Pix)/4; i++ {
			a := uint32(rgba.Pix[i*4+3])
			b := uint32(rgba.Pix[i*4+2])
			g := uint32(rgba.Pix[i*4+1])
			r := uint32(rgba.Pix[i*4+0])
			mm.Data[i] = ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | (a & 0xFF)
		}

		img.MipMaps = append(img.MipMaps, mm)
		if w > 4 {
			w = concepts.UMax(4, w/2)
		}
		if h > 4 {
			h = concepts.UMax(4, h/2)
		}
	}
}

func (img *Image) generateSimpleMipMaps() {
	img.MipMaps = make([]mipMap, 1)
	img.MipMaps[0] = mipMap{Width: img.Width, Height: img.Height, Data: img.Data}
	prev := img.MipMaps[0]

	w := img.Width / 2
	h := img.Height / 2

	var x, y, px, py, pcx, pcy uint32

	for w > 2 && h > 2 {
		mm := mipMap{Width: w, Height: h, Data: make([]uint32, w*h)}

		for y = 0; y < h; y++ {
			for x = 0; x < w; x++ {
				px = x * (prev.Width - 1) / (w - 1)
				py = y * (prev.Height - 1) / (h - 1)
				pcx = concepts.UMin(px+1, prev.Width-1)
				pcy = concepts.UMin(py+1, prev.Height-1)
				c := [16]color.RGBA{
					concepts.Int32ToRGBA(prev.Data[py*prev.Width+px]),
					concepts.Int32ToRGBA(prev.Data[py*prev.Width+pcx]),
					concepts.Int32ToRGBA(prev.Data[pcy*prev.Width+pcx]),
					concepts.Int32ToRGBA(prev.Data[pcy*prev.Width+px]),
				}
				avg := color.RGBA{
					uint8((uint32(c[0].R) + uint32(c[1].R) + uint32(c[2].R) + uint32(c[3].R)) / 4),
					uint8((uint32(c[0].G) + uint32(c[1].G) + uint32(c[2].G) + uint32(c[3].G)) / 4),
					uint8((uint32(c[0].B) + uint32(c[1].B) + uint32(c[2].B) + uint32(c[3].B)) / 4),
					uint8((uint32(c[0].A) + uint32(c[1].A) + uint32(c[2].A) + uint32(c[3].A)) / 4),
				}
				mm.Data[y*mm.Width+x] = concepts.RGBAToInt32(avg)
			}
		}
		img.MipMaps = append(img.MipMaps, mm)
		prev = mm
		if w > 2 {
			w = concepts.UMax(2, w/2)
		}
		if h > 2 {
			h = concepts.UMax(2, h/2)
		}
	}
}

func (img *Image) generateMipMaps() {
	img.MipMaps = make([]mipMap, 1)
	img.MipMaps[0] = mipMap{Width: img.Width, Height: img.Height, Data: img.Data}

	w := img.Width / 2
	h := img.Height / 2

	for w >= 2 && h >= 2 {
		mm := mipMap{Width: w, Height: h, Data: make([]uint32, w*h)}
		mm.Image = image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
		g := gift.New(
			gift.Resize(int(w), int(h), gift.LanczosResampling),
		)
		g.Draw(mm.Image, img.Image)
		for y := 0; y < int(h); y++ {
			for x := 0; x < int(w); x++ {
				index := x*4 + y*mm.Image.Stride
				mm.Data[y*int(mm.Width)+x] = concepts.RGBAToInt32(
					color.RGBA{mm.Image.Pix[index],
						mm.Image.Pix[index+1],
						mm.Image.Pix[index+2],
						mm.Image.Pix[index+3]})
			}
		}
		img.MipMaps = append(img.MipMaps, mm)

		w /= 2
		h /= 2
	}
}

func (img *Image) Sample(x, y float64, sw, sh uint32) concepts.Vector4 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := img.Data
	w := img.Width
	h := img.Height
	scaledArea := sw * sh

	if scaledArea > 0 && img.GenerateMipMaps && len(img.MipMaps) > 1 {
		mm := img.MipMaps[0]
		for i := 1; i < len(img.MipMaps); i++ {
			next := img.MipMaps[i]
			if scaledArea <= next.Width*next.Height {
				mm = next
				continue
			}
			data = mm.Data
			w = mm.Width
			h = mm.Height
			break
		}
	}

	if data == nil || w == 0 || h == 0 {
		// Debug values
		return concepts.Vector4{x, y, 0, 1} // full alpha
	}

	if x < 0 || y < 0 || x >= 1 || y >= 1 {
		return concepts.Vector4{0, 0, 0, 0}
	}

	fx := uint32(x * float64(w))
	fy := uint32(y * float64(h))

	if !img.Filter {
		// TODO: avoid allocating a vector here.
		c := data[fy*w+fx]
		return concepts.Vector4{
			float64((c>>24)&0xFF) / 255.0, float64((c>>16)&0xFF) / 255.0,
			float64((c>>8)&0xFF) / 255.0, float64(c&0xFF) / 255.0}
	}

	fx = concepts.UMin(fx, w-1)
	fy = concepts.UMin(fy, h-1)
	cx := (fx + 1) % w
	cy := (fy + 1) % h
	t00 := data[fy*w+fx]
	t10 := data[fy*w+cx]
	t11 := data[cy*w+cx]
	t01 := data[cy*w+fx]
	wx := x*float64(w) - float64(fx)
	wy := y*float64(h) - float64(fy)

	var r, g, b, a float64
	c00 := (t00 >> 24) & 0xFF
	c10 := (t10 >> 24) & 0xFF
	c11 := (t11 >> 24) & 0xFF
	c01 := (t01 >> 24) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		r = float64(c00)
	} else {
		r = float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy
	}
	c00 = (t00 >> 16) & 0xFF
	c10 = (t10 >> 16) & 0xFF
	c11 = (t11 >> 16) & 0xFF
	c01 = (t01 >> 16) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		g = float64(c00)
	} else {
		g = float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy
	}
	c00 = (t00 >> 8) & 0xFF
	c10 = (t10 >> 8) & 0xFF
	c11 = (t11 >> 8) & 0xFF
	c01 = (t01 >> 8) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		b = float64(c00)
	} else {
		b = float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy
	}
	c00 = t00 & 0xFF
	c10 = t10 & 0xFF
	c11 = t11 & 0xFF
	c01 = t01 & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		a = float64(c00)
	} else {
		a = float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy
	}
	a /= 255.0
	r = r / 255.0
	g = g / 255.0
	b = b / 255.0
	return concepts.Vector4{r, g, b, a}
}

func (img *Image) SampleAlpha(x, y float64, sw, sh uint32) float64 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := img.Data
	w := img.Width
	h := img.Height
	scaledArea := sw * sh

	if scaledArea > 0 && img.GenerateMipMaps && len(img.MipMaps) > 1 {
		mm := img.MipMaps[0]
		for i := 1; i < len(img.MipMaps); i++ {
			next := img.MipMaps[i]
			if scaledArea <= next.Width*next.Height {
				mm = next
				continue
			}
			data = mm.Data
			w = mm.Width
			h = mm.Height
			break
		}
	}

	if data == nil || w == 0 || h == 0 {
		// Debug values
		return 1 // full alpha
	}

	if x < 0 || y < 0 || x >= 1 || y >= 1 {
		return 0
	}

	fx := uint32(x * float64(w))
	fy := uint32(y * float64(h))

	if !img.Filter {
		return float64(data[fy*w+fx]&0xFF) / 255.0
	}

	fx = concepts.UMin(fx, w-1)
	fy = concepts.UMin(fy, h-1)
	cx := (fx + 1) % w
	cy := (fy + 1) % h
	wx := x*float64(w) - float64(fx)
	wy := y*float64(h) - float64(fy)

	c00 := data[fy*w+fx] & 0xFF
	c10 := data[fy*w+cx] & 0xFF
	c11 := data[cy*w+cx] & 0xFF
	c01 := data[cy*w+fx] & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		return float64(c00) / 255.0
	} else {
		return (float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy) / 255.0
	}
}

func (img *Image) Construct(data map[string]any) {
	img.Attached.Construct(data)
	img.Filter = false
	img.GenerateMipMaps = true

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		img.Source = v.(string)
	}
	if v, ok := data["GenerateMipMaps"]; ok {
		img.GenerateMipMaps = v.(bool)
	}
	if v, ok := data["Filter"]; ok {
		img.Filter = v.(bool)
	}
	img.Load()
}

func (img *Image) Serialize() map[string]any {
	result := img.Attached.Serialize()
	result["Source"] = img.Source
	result["GenerateMipMaps"] = img.GenerateMipMaps
	result["Filter"] = img.Filter
	return result
}
