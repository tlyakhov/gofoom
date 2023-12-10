package materials

import (
	"image"
	"image/color"
	"os"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"tlyakhov/gofoom/concepts"
)

type mipMap struct {
	Width, Height uint32
	Data          []uint32
}

// Image represents an image that will be rendered in-game.
type Image struct {
	concepts.Attached `editable:"^"`

	Width, Height   uint32
	Source          string `editable:"Texture Source" edit_type:"file"`
	GenerateMipMaps bool   `editable:"Generate Mip Maps?" edit_type:"bool"`
	Filter          bool   `editable:"Filter?" edit_type:"bool"`
	Data            []uint32
	MipMaps         map[uint32]*mipMap
	SmallestMipMap  *mipMap
}

var ImageComponentIndex int

func init() {
	ImageComponentIndex = concepts.DbTypes().Register(Image{}, ImageFromDb)
}

func ImageFromDb(entity *concepts.EntityRef) *Image {
	if asserted, ok := entity.Component(ImageComponentIndex).(*Image); ok {
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
	decoded, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Let's convert to 0-based NRGBA for speed/convenience.
	bounds := decoded.Bounds()
	img.Width = uint32(bounds.Dx())
	img.Height = uint32(bounds.Dy())
	img.Data = make([]uint32, int(img.Width)*int(img.Height))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			index := uint32(x-bounds.Min.X) + uint32(y-bounds.Min.Y)*img.Width
			// Premultiplied alpha
			img.Data[index] = concepts.ColorToInt32(decoded.At(x, y))
		}
	}
	img.generateMipMaps()
	return nil
}

func (img *Image) generateMipMaps() {
	img.MipMaps = make(map[uint32]*mipMap)

	index := concepts.NearestPow2(uint32(img.Height))
	img.MipMaps[index] = &mipMap{Width: img.Width, Height: img.Height, Data: img.Data}
	prev := img.MipMaps[index]

	w := img.Width / 2
	h := img.Height / 2

	var x, y, px, py, pcx, pcy uint32

	for w > 1 && h > 1 {
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
		index := concepts.NearestPow2(uint32(mm.Height))
		img.MipMaps[index] = &mm
		prev = &mm
		if w > 1 {
			w = concepts.UMax(1, w/2)
		}
		if h > 1 {
			h = concepts.UMax(1, h/2)
		}
	}
	img.SmallestMipMap = prev
}

func (img *Image) Sample(x, y float64, scale float64) concepts.Vector4 {
	// Testing:
	// return (0xAF << 24) | 0xFF
	data := img.Data
	w := img.Width
	h := img.Height
	scaledHeight := uint32(float64(h) * scale)

	if scaledHeight > 0 && img.GenerateMipMaps && img.SmallestMipMap != nil {
		if scaledHeight < img.SmallestMipMap.Height {
			data = img.SmallestMipMap.Data
			w = img.SmallestMipMap.Width
			h = img.SmallestMipMap.Height
		} else if scaledHeight < img.Height {
			mm := img.MipMaps[concepts.NearestPow2(scaledHeight)]
			data = mm.Data
			w = mm.Width
			h = mm.Height
		}
	}

	if data == nil || w == 0 || h == 0 {
		// Debug values
		return concepts.Vector4{x, y, 0, 1} // full alpha
	}

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	fx := uint32(x * float64(w))
	fy := uint32(y * float64(h))

	if !img.Filter {
		index := concepts.UMin(fy, h-1)*w + concepts.UMin(fx, w-1)
		//r := concepts.Vector4{}
		//concepts.AsmInt32ToVector4PreMul(data[index], (*[4]float64)(&r))
		//return r
		return concepts.Int32ToVector4PreMul(data[index])
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
