// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package materials

import (
	"image"
	"os"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

type ImageMipMap struct {
	Width, Height uint32
	PixelsLinear  []concepts.Vector4
	Image         *image.RGBA
}

// Image represents an image that will be rendered in-game.
type Image struct {
	ecs.Attached `editable:"^"`

	Width, Height   uint32
	Source          string `editable:"File" edit_type:"file"`
	GenerateMipMaps bool   `editable:"Generate Mip Maps?" edit_type:"bool"`
	Filter          bool   `editable:"Filter?" edit_type:"bool"`
	ConvertSRGB     bool   `editable:"sRGB->Linear?"`
	PixelsRGBA      []uint32
	PixelsLinear    []concepts.Vector4
	MipMaps         []ImageMipMap
	Image           image.Image
}

var ImageCID ecs.ComponentID

func init() {
	ImageCID = ecs.RegisterComponent(&ecs.Column[Image, *Image]{Getter: GetImage})
}

func GetImage(u *ecs.Universe, e ecs.Entity) *Image {
	if asserted, ok := u.Component(e, ImageCID).(*Image); ok {
		return asserted
	}
	return nil
}

func (img *Image) MultiAttachable() bool { return true }

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
	img.PixelsRGBA = make([]uint32, int(img.Width)*int(img.Height))
	img.PixelsLinear = make([]concepts.Vector4, len(img.PixelsRGBA))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			index := uint32(x-bounds.Min.X) + uint32(y-bounds.Min.Y)*img.Width
			// Premultiplied alpha
			img.PixelsRGBA[index] = concepts.ColorToInt32PreMul(img.Image.At(x, y))
		}
	}
	return nil
}

func (img *Image) Sample(x, y float64, sw, sh uint32) concepts.Vector4 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := img.PixelsLinear
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
			data = mm.PixelsLinear
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
		return data[fy*w+fx]
	}

	fx = concepts.Min(fx, w-1)
	fy = concepts.Min(fy, h-1)
	cx := (fx + 1) % w
	cy := (fy + 1) % h
	t00 := data[fy*w+fx]
	t10 := data[fy*w+cx]
	t11 := data[cy*w+cx]
	t01 := data[cy*w+fx]
	wx := x*float64(w) - float64(fx)
	wy := y*float64(h) - float64(fy)

	var r, g, b, a float64
	c00 := t00[0]
	c10 := t10[0]
	c11 := t11[0]
	c01 := t01[0]
	if c00 == c10 && c10 == c11 && c11 == c01 {
		r = c00
	} else {
		r = c00*(1.0-wx)*(1.0-wy) + c10*wx*(1.0-wy) + c11*wx*wy + c01*(1.0-wx)*wy
	}
	c00 = t00[1]
	c10 = t10[1]
	c11 = t11[1]
	c01 = t01[1]
	if c00 == c10 && c10 == c11 && c11 == c01 {
		g = c00
	} else {
		g = c00*(1.0-wx)*(1.0-wy) + c10*wx*(1.0-wy) + c11*wx*wy + c01*(1.0-wx)*wy
	}
	c00 = t00[2]
	c10 = t10[2]
	c11 = t11[2]
	c01 = t01[2]
	if c00 == c10 && c10 == c11 && c11 == c01 {
		b = c00
	} else {
		b = c00*(1.0-wx)*(1.0-wy) + c10*wx*(1.0-wy) + c11*wx*wy + c01*(1.0-wx)*wy
	}
	c00 = t00[3]
	c10 = t10[3]
	c11 = t11[3]
	c01 = t01[3]
	if c00 == c10 && c10 == c11 && c11 == c01 {
		a = c00
	} else {
		a = c00*(1.0-wx)*(1.0-wy) + c10*wx*(1.0-wy) + c11*wx*wy + c01*(1.0-wx)*wy
	}
	return concepts.Vector4{r, g, b, a}
}

func (img *Image) SampleAlpha(x, y float64, sw, sh uint32) float64 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := img.PixelsLinear
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
			data = mm.PixelsLinear
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
		return data[fy*w+fx][3]
	}

	fx = concepts.Min(fx, w-1)
	fy = concepts.Min(fy, h-1)
	cx := (fx + 1) % w
	cy := (fy + 1) % h
	wx := x*float64(w) - float64(fx)
	wy := y*float64(h) - float64(fy)

	c00 := data[fy*w+fx][3]
	c10 := data[fy*w+cx][3]
	c11 := data[cy*w+cx][3]
	c01 := data[cy*w+fx][3]
	if c00 == c10 && c10 == c11 && c11 == c01 {
		return c00
	} else {
		return c00*(1.0-wx)*(1.0-wy) + c10*wx*(1.0-wy) + c11*wx*wy + c01*(1.0-wx)*wy
	}
}

func (img *Image) Construct(data map[string]any) {
	img.Attached.Construct(data)
	img.Filter = false
	img.GenerateMipMaps = true
	img.ConvertSRGB = true

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		img.Source = v.(string)
	}
	if v, ok := data["GenerateMipMaps"]; ok {
		img.GenerateMipMaps = cast.ToBool(v)
	}
	if v, ok := data["Filter"]; ok {
		img.Filter = cast.ToBool(v)
	}
	if v, ok := data["ConvertSRGB"]; ok {
		img.ConvertSRGB = cast.ToBool(v)
	}
	img.Load()
}

func (img *Image) Serialize() map[string]any {
	result := img.Attached.Serialize()
	result["Source"] = img.Source
	result["Filter"] = img.Filter
	if !img.GenerateMipMaps {
		result["GenerateMipMaps"] = img.GenerateMipMaps
	}
	if !img.ConvertSRGB {
		result["ConvertSRGB"] = img.ConvertSRGB
	}
	return result
}
