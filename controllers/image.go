// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"image"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/disintegration/gift"
)

type ImageController struct {
	ecs.BaseController
	*materials.Image
	toneMap *materials.ToneMap
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ImageController{} }, 100)
}

func (ic *ImageController) ComponentID() ecs.ComponentID {
	return materials.ImageCID
}

func (ic *ImageController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (ic *ImageController) Target(target ecs.Attachable, e ecs.Entity) bool {
	ic.Entity = e
	ic.Image = target.(*materials.Image)
	return ic.IsActive()
}

func (ic *ImageController) Recalculate() {
	if ic.toneMap == nil {
		ic.toneMap = ecs.Singleton(materials.ToneMapCID).(*materials.ToneMap)
	}

	for y := 0; y < int(ic.Height); y++ {
		for x := 0; x < int(ic.Width); x++ {
			index := x + y*int(ic.Width)
			if !ic.ConvertSRGB {
				ic.PixelsLinear[index] = concepts.Vector4{
					float64((ic.PixelsRGBA[index]>>24)&0xFF) / 255,
					float64((ic.PixelsRGBA[index]>>16)&0xFF) / 255,
					float64((ic.PixelsRGBA[index]>>8)&0xFF) / 255,
					float64(ic.PixelsRGBA[index]&0xFF) / 255,
				}
				continue
			}
			ic.PixelsLinear[index] = concepts.Vector4{
				ic.toneMap.LutSRGBToLinear[(ic.PixelsRGBA[index]>>24)&0xFF],
				ic.toneMap.LutSRGBToLinear[(ic.PixelsRGBA[index]>>16)&0xFF],
				ic.toneMap.LutSRGBToLinear[(ic.PixelsRGBA[index]>>8)&0xFF],
				float64(ic.PixelsRGBA[index]&0xFF) / 255,
			}
		}
	}

	if ic.GenerateMipMaps {
		ic.generateMipMaps()
	}
}

/*func (img *ImageController) generateTestMipMaps() {
	img.MipMaps = make([]materials.ImageMipMap, 0)
	w := img.Width
	h := img.Height
	for w > 4 && h > 4 {
		index := len(img.MipMaps)
		bg := (index + 1) * 255 / 6
		mm := materials.ImageMipMap{Width: w, Height: h, PixelsLinear: make([]concepts.Vector4, w*h)}
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
			mm.PixelsLinear[i] = ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | (a & 0xFF)
		}

		img.MipMaps = append(img.MipMaps, mm)
		if w > 4 {
			w = concepts.Max(4, w/2)
		}
		if h > 4 {
			h = concepts.Max(4, h/2)
		}
	}
}

func (img *ImageController) generateSimpleMipMaps() {
	img.MipMaps = make([]materials.ImageMipMap, 1)
	img.MipMaps[0] = materials.ImageMipMap{Width: img.Width, Height: img.Height, PixelsLinear: img.PixelsRGBA}
	prev := img.MipMaps[0]

	w := img.Width / 2
	h := img.Height / 2

	var x, y, px, py, pcx, pcy uint32

	for w > 2 && h > 2 {
		mm := materials.ImageMipMap{Width: w, Height: h, PixelsLinear: make([]uint32, w*h)}

		for y = 0; y < h; y++ {
			for x = 0; x < w; x++ {
				px = x * (prev.Width - 1) / (w - 1)
				py = y * (prev.Height - 1) / (h - 1)
				pcx = concepts.Min(px+1, prev.Width-1)
				pcy = concepts.Min(py+1, prev.Height-1)
				c := [16]color.RGBA{
					concepts.Int32ToRGBA(prev.PixelsLinear[py*prev.Width+px]),
					concepts.Int32ToRGBA(prev.PixelsLinear[py*prev.Width+pcx]),
					concepts.Int32ToRGBA(prev.PixelsLinear[pcy*prev.Width+pcx]),
					concepts.Int32ToRGBA(prev.PixelsLinear[pcy*prev.Width+px]),
				}
				avg := color.RGBA{
					uint8((uint32(c[0].R) + uint32(c[1].R) + uint32(c[2].R) + uint32(c[3].R)) / 4),
					uint8((uint32(c[0].G) + uint32(c[1].G) + uint32(c[2].G) + uint32(c[3].G)) / 4),
					uint8((uint32(c[0].B) + uint32(c[1].B) + uint32(c[2].B) + uint32(c[3].B)) / 4),
					uint8((uint32(c[0].A) + uint32(c[1].A) + uint32(c[2].A) + uint32(c[3].A)) / 4),
				}
				mm.PixelsLinear[y*mm.Width+x] = concepts.RGBAToInt32(avg)
			}
		}
		img.MipMaps = append(img.MipMaps, mm)
		prev = mm
		if w > 2 {
			w = concepts.Max(2, w/2)
		}
		if h > 2 {
			h = concepts.Max(2, h/2)
		}
	}
}*/

func (ic *ImageController) generateMipMaps() {
	ic.MipMaps = make([]materials.ImageMipMap, 1)
	ic.MipMaps[0] = materials.ImageMipMap{Width: ic.Width, Height: ic.Height, PixelsLinear: ic.PixelsLinear}

	w := ic.Width / 2
	h := ic.Height / 2

	for w >= 2 && h >= 2 {
		mm := materials.ImageMipMap{Width: w, Height: h, PixelsLinear: make([]concepts.Vector4, w*h)}
		mm.Image = image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
		g := gift.New(
			gift.Resize(int(w), int(h), gift.LanczosResampling),
		)
		g.Draw(mm.Image, ic.Image.Image)
		for y := 0; y < int(h); y++ {
			for x := 0; x < int(w); x++ {
				index := x*4 + y*mm.Image.Stride
				if !ic.ConvertSRGB {
					mm.PixelsLinear[y*int(mm.Width)+x] = concepts.Vector4{
						float64(mm.Image.Pix[index]) / 255,
						float64(mm.Image.Pix[index+1]) / 255,
						float64(mm.Image.Pix[index+2]) / 255,
						float64(mm.Image.Pix[index+3]) / 255,
					}
					continue
				}
				mm.PixelsLinear[y*int(mm.Width)+x] = concepts.Vector4{
					ic.toneMap.LutSRGBToLinear[mm.Image.Pix[index]],
					ic.toneMap.LutSRGBToLinear[mm.Image.Pix[index+1]],
					ic.toneMap.LutSRGBToLinear[mm.Image.Pix[index+2]],
					float64(mm.Image.Pix[index+3]) / 255,
				}
			}
		}
		ic.MipMaps = append(ic.MipMaps, mm)

		w /= 2
		h /= 2
	}
}
