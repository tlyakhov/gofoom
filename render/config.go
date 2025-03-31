// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"math"
	"sync"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type Config struct {
	Universe                  *ecs.Universe
	Multithreaded             bool
	NumBlocks                 int
	ScreenWidth, ScreenHeight int
	MaxViewDist, FOV          float64
	LightGrid                 float64
	CameraToProjectionPlane   float64
	ViewRadians               []float64
	ViewFix                   []float64
	ZBuffer                   []float64
	FrameBuffer               []concepts.Vector4
	// For walls over portals
	ExtraBuffer []concepts.Vector4
	FrameTint   concepts.Vector4
	Player      *behaviors.Player
	PlayerBody  *core.Body
	Carrier     *behaviors.InventoryCarrier

	RenderLock sync.Mutex
}

func (c *Config) Initialize() {
	c.CameraToProjectionPlane = (float64(c.ScreenWidth) / 2.0) / math.Tan(c.FOV*concepts.Deg2rad/2.0)
	c.ViewRadians = make([]float64, c.ScreenWidth)
	c.ViewFix = make([]float64, c.ScreenWidth)

	for i := 0; i < c.ScreenWidth; i++ {
		// See https://stackoverflow.com/questions/24173966/raycasting-engine-rendering-creating-slight-distortion-increasing-towards-edges
		c.ViewRadians[i] = math.Atan(float64(i-c.ScreenWidth/2) / c.CameraToProjectionPlane)
		c.ViewFix[i] = c.CameraToProjectionPlane / math.Cos(c.ViewRadians[i])
	}

	c.ZBuffer = make([]float64, c.ScreenWidth*c.ScreenHeight)
	c.FrameBuffer = make([]concepts.Vector4, c.ScreenWidth*c.ScreenHeight)
	c.ExtraBuffer = make([]concepts.Vector4, c.ScreenWidth*c.ScreenHeight)

	c.RefreshPlayer()
}

func (c *Config) RefreshPlayer() {
	col := ecs.ColumnFor[behaviors.Player](c.Universe, behaviors.PlayerCID)
	for i := range col.Cap() {
		player := col.Value(i)
		if player == nil || !player.IsActive() || player.Spawn {
			continue
		}
		c.Player = player
		c.PlayerBody = core.GetBody(c.Universe, player.Entity)
		c.Carrier = behaviors.GetInventoryCarrier(c.Universe, player.Entity)
		return
	}
}

const lightmapVMask uint64 = (1 << 16) - 1
const lightmapNMask uint64 = (1 << 5) - 1

func (c *Config) WorldToLightmapHash(s *core.Sector, v *concepts.Vector3, n *concepts.Vector3) uint64 {
	// Floor is important, needs to truncate towards -Infinity rather than 0
	z := int64(math.Floor(v[2]/c.LightGrid)) - s.LightmapBias[2]
	y := int64(math.Floor(v[1]/c.LightGrid)) - s.LightmapBias[1]
	x := int64(math.Floor(v[0]/c.LightGrid)) - s.LightmapBias[0]
	// Quantize the normal to 5 bits per dimension
	nz := int8(n[2] * 15)
	ny := int8(n[1] * 15)
	nx := int8(n[0] * 15)
	// Bit shift and mask the components
	return ((uint64(x) & lightmapVMask) << 48) |
		((uint64(y) & lightmapVMask) << 32) |
		((uint64(z) & lightmapVMask) << 16) |
		((uint64(nx) & lightmapNMask) << 10) |
		((uint64(ny) & lightmapNMask) << 5) |
		((uint64(nz) & lightmapNMask) << 0)
}

func (c *Config) LightmapHashToWorld(s *core.Sector, result *concepts.Vector3, a uint64) *concepts.Vector3 {
	// Bypass the normal
	a = a >> 16
	z := int64(int16(a&lightmapVMask)) + s.LightmapBias[2]
	result[2] = float64(z * int64(c.LightGrid))
	a = a >> 16
	y := int64(int16(a&lightmapVMask)) + s.LightmapBias[1]
	result[1] = float64(y * int64(c.LightGrid))
	a = a >> 16
	x := int64(int16(a&lightmapVMask)) + s.LightmapBias[0]
	result[0] = float64(x * int64(c.LightGrid))
	return result
}
