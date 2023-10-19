package texture

import (
	"image"
	"image/color"
	"os"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type mipMap struct {
	Width, Height uint32
	Data          []uint32
}

// Image represents an image that will be rendered in-game.
type Image struct {
	concepts.Base

	Width, Height   uint32
	Source          string `editable:"Texture Source" edit_type:"string"`
	GenerateMipMaps bool   `editable:"Generate Mip Maps?" edit_type:"bool"`
	Filter          bool   `editable:"Filter?" edit_type:"bool"`
	Data            []uint32
	MipMaps         map[uint32]*mipMap
	SmallestMipMap  *mipMap
}

func init() {
	registry.Instance().Register(Image{})
}

func (t *Image) Initialize() {
	t.Base = concepts.Base{}
	t.Base.Initialize()
	t.Filter = false
	t.GenerateMipMaps = true
}

// Load a texture from a file (pre-processing mipmaps if set)
func (t *Image) Load() error {
	if t.Source == "" {
		return nil
	}

	// Load the image from a file...
	file, err := os.Open(t.Source)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Let's convert to 0-based NRGBA for speed/convenience.
	bounds := img.Bounds()
	t.Width = uint32(bounds.Dx())
	t.Height = uint32(bounds.Dy())
	t.Data = make([]uint32, int(t.Width)*int(t.Height))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			index := uint32(x-bounds.Min.X) + uint32(y-bounds.Min.Y)*t.Width
			t.Data[index] = concepts.ColorToInt32(img.At(x, y))
		}
	}
	t.generateMipMaps()
	return nil
}

func (t *Image) generateMipMaps() {
	t.MipMaps = make(map[uint32]*mipMap)

	index := concepts.NearestPow2(uint32(t.Height))
	t.MipMaps[index] = &mipMap{Width: t.Width, Height: t.Height, Data: t.Data}
	prev := t.MipMaps[index]

	w := t.Width / 2
	h := t.Height / 2

	var x, y, px, py, pcx, pcy uint32

	for w > 1 && h > 1 {
		mm := mipMap{Width: w, Height: h, Data: make([]uint32, w*h)}

		for y = 0; y < h; y++ {
			for x = 0; x < w; x++ {
				px = x * (prev.Width - 1) / (w - 1)
				py = y * (prev.Height - 1) / (h - 1)
				pcx = concepts.UMin(px+1, prev.Width-1)
				pcy = concepts.UMin(py+1, prev.Height-1)
				c := [16]color.NRGBA{
					concepts.Int32ToNRGBA(prev.Data[py*prev.Width+px]),
					concepts.Int32ToNRGBA(prev.Data[py*prev.Width+pcx]),
					concepts.Int32ToNRGBA(prev.Data[pcy*prev.Width+pcx]),
					concepts.Int32ToNRGBA(prev.Data[pcy*prev.Width+px]),
				}
				avg := color.NRGBA{
					uint8((uint32(c[0].R) + uint32(c[1].R) + uint32(c[2].R) + uint32(c[3].R)) / 4),
					uint8((uint32(c[0].G) + uint32(c[1].G) + uint32(c[2].G) + uint32(c[3].G)) / 4),
					uint8((uint32(c[0].B) + uint32(c[1].B) + uint32(c[2].B) + uint32(c[3].B)) / 4),
					uint8((uint32(c[0].A) + uint32(c[1].A) + uint32(c[2].A) + uint32(c[3].A)) / 4),
				}
				mm.Data[y*mm.Width+x] = concepts.NRGBAToInt32(avg)
			}
		}
		index := concepts.NearestPow2(uint32(mm.Height))
		t.MipMaps[index] = &mm
		prev = &mm
		if w > 1 {
			w = concepts.UMax(1, w/2)
		}
		if h > 1 {
			h = concepts.UMax(1, h/2)
		}
	}
	t.SmallestMipMap = prev
}

func (t *Image) Sample(x, y float64, scale float64) uint32 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := t.Data
	w := t.Width
	h := t.Height
	scaledHeight := uint32(float64(h) * scale)

	if scaledHeight > 0 && t.GenerateMipMaps && t.SmallestMipMap != nil {
		if scaledHeight < t.SmallestMipMap.Height {
			data = t.SmallestMipMap.Data
			w = t.SmallestMipMap.Width
			h = t.SmallestMipMap.Height
		} else if scaledHeight < t.Height {
			mm := t.MipMaps[concepts.NearestPow2(scaledHeight)]
			data = mm.Data
			w = mm.Width
			h = mm.Height
		}
	}

	if data == nil || w == 0 || h == 0 {
		return 0xFF // Black, full alpha
	}

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	fx := uint32(x * float64(w))
	fy := uint32(y * float64(h))

	if !t.Filter {
		index := concepts.UMin(fy, h-1)*w + concepts.UMin(fx, w-1)
		return data[index]
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

	var r, g, b, a uint32
	c00 := (t00 >> 24) & 0xFF
	c10 := (t10 >> 24) & 0xFF
	c11 := (t11 >> 24) & 0xFF
	c01 := (t01 >> 24) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		r = c00
	} else {
		r = uint32(float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy)
	}
	c00 = (t00 >> 16) & 0xFF
	c10 = (t10 >> 16) & 0xFF
	c11 = (t11 >> 16) & 0xFF
	c01 = (t01 >> 16) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		g = c00
	} else {
		g = uint32(float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy)
	}
	c00 = (t00 >> 8) & 0xFF
	c10 = (t10 >> 8) & 0xFF
	c11 = (t11 >> 8) & 0xFF
	c01 = (t01 >> 8) & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		b = c00
	} else {
		b = uint32(float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy)
	}
	c00 = t00 & 0xFF
	c10 = t10 & 0xFF
	c11 = t11 & 0xFF
	c01 = t01 & 0xFF
	if c00 == c10 && c10 == c11 && c11 == c01 {
		a = c00
	} else {
		a = uint32(float64(c00)*(1.0-wx)*(1.0-wy) + float64(c10)*wx*(1.0-wy) + float64(c11)*wx*wy + float64(c01)*(1.0-wx)*wy)
	}
	return ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | (a & 0xFF)
}

func (t *Image) Deserialize(data map[string]interface{}) {
	t.Initialize()
	t.Base.Deserialize(data)
	if v, ok := data["Source"]; ok {
		t.Source = v.(string)
	}
	if v, ok := data["GenerateMipMaps"]; ok {
		t.GenerateMipMaps = v.(bool)
	}
	if v, ok := data["Filter"]; ok {
		t.Filter = v.(bool)
	}
	t.Load()
}

func (t *Image) Serialize() map[string]interface{} {
	result := t.Base.Serialize()
	result["Type"] = "texture.Image"
	result["Source"] = t.Source
	result["GenerateMipMaps"] = t.GenerateMipMaps
	result["Filter"] = t.Filter
	return result
}
