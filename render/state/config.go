package state

import (
	"math"
	"sync"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
)

type Config struct {
	DB                        *concepts.EntityComponentDB
	ScreenWidth, ScreenHeight int
	Frame, Counter            int
	MaxViewDist, FOV          float64
	CameraToProjectionPlane   float64
	ViewRadians               []float64
	ViewFix                   []float64
	ZBuffer                   []float64
	DebugNotices              concepts.SyncUniqueQueue
	RenderLock                sync.Mutex
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
}

// Player is a convenience function to get the player this renderer links to.
func (c *Config) Player() *behaviors.Player {
	return c.DB.First(behaviors.PlayerComponentIndex).(*behaviors.Player)
}
