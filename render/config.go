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
	ECS                       *ecs.ECS
	Multithreaded             bool
	Blocks                    int
	ScreenWidth, ScreenHeight int
	MaxViewDist, FOV          float64
	LightGrid                 float64
	CameraToProjectionPlane   float64
	ViewRadians               []float64
	ViewFix                   []float64
	ZBuffer                   []float64
	FrameBuffer               []concepts.Vector4
	FrameTint                 [4]float64
	Player                    *behaviors.Player
	PlayerBody                *core.Body

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

	c.RefreshPlayer()
}

func (c *Config) RefreshPlayer() {
	col := ecs.ColumnFor[behaviors.Player](c.ECS, behaviors.PlayerCID)
	for i := range col.Cap() {
		player := col.Value(i)
		if player == nil || !player.Active || player.Spawn {
			continue
		}
		c.Player = player
		c.PlayerBody = core.GetBody(c.ECS, c.Player.Entity)
		return
	}
}

const lightmapMask uint64 = (1 << 16) - 1

func (c *Config) WorldToLightmapAddress(s *core.Sector, v *concepts.Vector3, extraHash uint16) uint64 {
	// Floor is important, needs to truncate towards -Infinity rather than 0
	z := int64(math.Floor(v[2]/c.LightGrid)) - s.LightmapBias[2]
	y := int64(math.Floor(v[1]/c.LightGrid)) - s.LightmapBias[1]
	x := int64(math.Floor(v[0]/c.LightGrid)) - s.LightmapBias[0]
	/*if x < 0 || y < 0 || z < 0 {
		fmt.Printf("Error: lightmap address conversion resulted in negative value: %v,%v,%v\n", x, y, z)
	}*/
	// Bit shift and mask the components, and add the sector entity at the end
	// to ensure that overlapping addresses are distinct for each sector
	return (((uint64(x) & lightmapMask) << 48) |
		((uint64(y) & lightmapMask) << 32) |
		((uint64(z) & lightmapMask) << 16) |
		uint64(extraHash)) //+ (uint64(s.Entity) * 1009)
}

func (c *Config) LightmapAddressToWorld(s *core.Sector, result *concepts.Vector3, a uint64) *concepts.Vector3 {
	//w := uint64(a & wMask)
	//a -= uint64(s.Entity) * 1009
	a = a >> 16
	z := int64((a & lightmapMask)) + s.LightmapBias[2]
	result[2] = float64(z) * c.LightGrid
	a = a >> 16
	y := int64((a & lightmapMask)) + s.LightmapBias[1]
	result[1] = float64(y) * c.LightGrid
	a = a >> 16
	x := int64((a & lightmapMask)) + s.LightmapBias[0]
	result[0] = float64(x) * c.LightGrid
	return result
}
