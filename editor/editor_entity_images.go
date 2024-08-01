// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"image"
	"image/color"
	"math"
	"time"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"

	"github.com/fogleman/gg"
)

func (e *Editor) materialSelectionBorderColor(entity concepts.Entity) *concepts.Vector4 {
	if materials.ShaderFromDb(e.DB, entity) != nil {
		return &concepts.Vector4{1.0, 0.0, 1.0, 0.5}
	}
	return &concepts.Vector4{0.0, 0.0, 0.0, 0.0}
}

func (e *Editor) imageForMaterial(entity concepts.Entity) image.Image {
	w, h := 64, 64
	e.MaterialSampler.Initialize(entity, nil)
	e.MaterialSampler.ScaleW = 64
	e.MaterialSampler.ScaleH = 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buffer := img.Pix
	border := e.materialSelectionBorderColor(entity)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			e.MaterialSampler.ScreenX = x * e.MaterialSampler.ScreenWidth / w
			e.MaterialSampler.ScreenY = y * e.MaterialSampler.ScreenHeight / h
			e.MaterialSampler.Angle = float64(x) * math.Pi * 2.0 / float64(w)
			e.MaterialSampler.SampleMaterial(nil, float64(x)/float64(w), float64(y)/float64(h))
			if e.MaterialSampler.NoTexture {
				return e.noTextureImage
			}
			if x <= 1 || y <= 1 || x >= w-2 || y >= h-2 {
				e.MaterialSampler.Output.AddPreMulColorSelf(border)
			}
			index := x*4 + y*img.Stride

			buffer[index+3] = concepts.ByteClamp(e.MaterialSampler.Output[3] * 0xFF)
			buffer[index+2] = concepts.ByteClamp(e.MaterialSampler.Output[2] * 0xFF)
			buffer[index+1] = concepts.ByteClamp(e.MaterialSampler.Output[1] * 0xFF)
			buffer[index+0] = concepts.ByteClamp(e.MaterialSampler.Output[0] * 0xFF)
		}
	}
	return img
}

var patternPrimary = gg.NewSolidPattern(color.NRGBA{255, 255, 255, 255})
var patternSecondary = gg.NewSolidPattern(color.NRGBA{255, 255, 0, 255})

func (e *Editor) imageForSector(entity concepts.Entity) image.Image {
	w, h := 64, 64
	context := gg.NewContext(w, h)

	sector := core.SectorFromDb(e.DB, entity)
	context.SetLineWidth(1)
	for _, segment := range sector.Segments {
		if segment.AdjacentSegment != nil {
			context.SetStrokeStyle(patternSecondary)
		} else {
			context.SetStrokeStyle(patternPrimary)
		}
		context.NewSubPath()
		x := (segment.P[0] - sector.Min[0]) * float64(w) / (sector.Max[0] - sector.Min[0])
		y := (segment.P[1] - sector.Min[1]) * float64(h) / (sector.Max[1] - sector.Min[1])
		context.MoveTo(x, y)
		x = (segment.Next.P[0] - sector.Min[0]) * float64(w) / (sector.Max[0] - sector.Min[0])
		y = (segment.Next.P[1] - sector.Min[1]) * float64(h) / (sector.Max[1] - sector.Min[1])
		context.LineTo(x, y)
		context.ClosePath()
		context.Stroke()
	}

	return context.Image()
}

func (e *Editor) EntityImage(entity concepts.Entity, sector bool) image.Image {
	item, exists := e.entityIconCache.Load(entity)
	now := time.Now().UnixMilli()
	if exists && now-item.LastUpdated < 1000*60 {
		return item.Image
	}
	item.Image = nil
	if sector {
		item.Image = e.imageForSector(entity)
	} else {
		item.Image = e.imageForMaterial(entity)
	}
	item.LastUpdated = now
	// TODO: Clean this cache periodically
	e.entityIconCache.Store(entity, item)
	return item.Image
}
