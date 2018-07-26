package engine

import (
	"image"
	"image/color"
	"math"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"os"

	"github.com/tlyakhov/gofoom/util"
)

type mipMap struct {
	Width, Height uint
	Data          *image.NRGBA
}

// Texture represents an image that will be rendered in-game.
type Texture struct {
	util.CommonFields

	Width, Height   uint
	Source          string `editable:"Texture Source" edit_type:"string"`
	GenerateMipMaps bool   `editable:"Generate Mip Maps?" edit_type:"bool"`
	Filter          bool   `editable:"Filter?" edit_type:"bool"`
	Data            *image.NRGBA
	MipMaps         map[uint]*mipMap
	SmallestMipMap  *mipMap
}

// Load a texture from a file (pre-processing mipmaps if set)
func (t *Texture) Load() (*Texture, error) {
	if t.Source == "" {
		return t, nil
	}

	// Load the image from a file...
	file, err := os.Open(t.Source)
	if err != nil {
		return t, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return t, err
	}

	// Let's convert to 0-based NRGBA for speed/convenience.
	bounds := img.Bounds()
	t.Width = uint(bounds.Dx())
	t.Height = uint(bounds.Dy())
	t.Data = image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(t.Width), int(t.Height)}})
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			t.Data.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}
	return t, nil
}

func (t *Texture) generateMipMaps() {
	t.MipMaps = make(map[uint]*mipMap)

	index := util.NearestPow2(uint(t.Height))
	t.MipMaps[index] = &mipMap{Width: t.Width, Height: t.Height, Data: t.Data}
	prev := t.MipMaps[index]

	w := t.Width / 2
	h := t.Height / 2

	var x, y, px, py, pcx, pcy uint

	for w > 1 && h > 1 {
		mm := mipMap{Width: w, Height: h, Data: image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(w), int(h)}})}

		for y = 0; y < h; y++ {
			for x = 0; x < w; x++ {
				px = x * (prev.Width - 1) / (w - 1)
				py = y * (prev.Height - 1) / (h - 1)
				pcx = util.UMin(px+1, prev.Width-1)
				pcy = util.UMin(py+1, prev.Height-1)
				c := [4]color.NRGBA{
					prev.Data.At(int(px), int(py)).(color.NRGBA),
					prev.Data.At(int(pcx), int(py)).(color.NRGBA),
					prev.Data.At(int(pcx), int(pcy)).(color.NRGBA),
					prev.Data.At(int(px), int(pcy)).(color.NRGBA),
				}
				avg := color.NRGBA{
					uint8((uint(c[0].R) + uint(c[1].R) + uint(c[2].R) + uint(c[3].R)) / 4),
					uint8((uint(c[0].G) + uint(c[1].G) + uint(c[2].G) + uint(c[3].G)) / 4),
					uint8((uint(c[0].B) + uint(c[1].B) + uint(c[2].B) + uint(c[3].B)) / 4),
					uint8((uint(c[0].A) + uint(c[1].A) + uint(c[2].A) + uint(c[3].A)) / 4),
				}
				mm.Data.Set(int(x), int(y), avg)
			}
		}
		index := util.NearestPow2(uint(mm.Height))
		t.MipMaps[index] = &mm
		prev = &mm
		if w > 1 {
			w = util.UMax(1, w/2)
		}
		if h > 1 {
			h = util.UMax(1, h/2)
		}
	}
	t.SmallestMipMap = prev
}

func (t *Texture) Sample(x, y float64, scaledHeight uint) color.NRGBA {
	data := t.Data
	w := t.Width
	h := t.Height

	if scaledHeight > 0 && t.GenerateMipMaps && t.SmallestMipMap != nil {
		if scaledHeight < t.SmallestMipMap.Height {
			data = t.SmallestMipMap.Data
			w = t.SmallestMipMap.Width
			h = t.SmallestMipMap.Height
		} else if scaledHeight < t.Height {
			mm := t.MipMaps[util.NearestPow2(scaledHeight)]
			data = mm.Data
			w = mm.Width
			h = mm.Height
		}
	}

	if data == nil || w == 0 || h == 0 {
		return color.NRGBA{0, 0, 0, 0xFF}
	}

	x = math.Max(x, 0)
	y = math.Max(y, 0)

	fx := uint(x * float64(w))
	fy := uint(y * float64(h))

	if !t.Filter {
		index := (util.UMin(fy, h)*w + util.UMin(fx, w)) * 4
		return color.NRGBA{data.Pix[index], data.Pix[index+1], data.Pix[index+2], data.Pix[index+3]}
	}

	wx := x - float64(fx)
	wy := y - float64(fy)
	fx = util.UMax(fx, w-1)
	fy = util.UMax(fy, h-1)
	cx := (fx + 1) % w
	cy := (fy + 1) % h
	t00 := (fy*w + fx) * 4
	t10 := (fy*w + cx) * 4
	t11 := (cy*w + cx) * 4
	t01 := (cy*w + fx) * 4
	return color.NRGBA{
		uint8(float64(data.Pix[t00])*(1.0-wx)*(1.0-wy) + float64(data.Pix[t10])*wx*(1.0-wy) + float64(data.Pix[t11])*wx*wy + float64(data.Pix[t01])*(1.0-wx)*wy),
		uint8(float64(data.Pix[t00+1])*(1.0-wx)*(1.0-wy) + float64(data.Pix[t10+1])*wx*(1.0-wy) + float64(data.Pix[t11+1])*wx*wy + float64(data.Pix[t01+1])*(1.0-wx)*wy),
		uint8(float64(data.Pix[t00+2])*(1.0-wx)*(1.0-wy) + float64(data.Pix[t10+2])*wx*(1.0-wy) + float64(data.Pix[t11+2])*wx*wy + float64(data.Pix[t01+2])*(1.0-wx)*wy),
		uint8(float64(data.Pix[t00+3])*(1.0-wx)*(1.0-wy) + float64(data.Pix[t10+3])*wx*(1.0-wy) + float64(data.Pix[t11+3])*wx*wy + float64(data.Pix[t01+3])*(1.0-wx)*wy),
	}
}
