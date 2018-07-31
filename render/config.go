package render

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
)

type trigEntry struct {
	sin, cos float64
}

type Config struct {
	ScreenWidth, ScreenHeight              int
	Frame, FrameTint, WorkerWidth, Counter int
	MaxViewDist, FOV                       float64
	CameraToProjectionPlane                float64
	TrigCount                              int
	TrigTable                              []trigEntry
	ViewFix                                []float64
	ZBuffer                                []float64
	FloorNormal                            concepts.Vector3
	CeilingNormal                          concepts.Vector3
}

func (c *Config) Initialize() {
	c.CameraToProjectionPlane = (float64(c.ScreenWidth) / 2.0) / math.Tan(c.FOV*concepts.Deg2rad/2.0)
	c.TrigCount = int(float64(c.ScreenWidth) * 360.0 / c.FOV) // Quantize trig tables per-Pixel.
	c.TrigTable = make([]trigEntry, c.TrigCount)
	c.ViewFix = make([]float64, c.ScreenWidth)

	for i := 0; i < c.TrigCount; i++ {
		c.TrigTable[i].sin = math.Sin(float64(i) * 2.0 * math.Pi / float64(c.TrigCount))
		c.TrigTable[i].cos = math.Cos(float64(i) * 2.0 * math.Pi / float64(c.TrigCount))
	}

	for i := 0; i < c.ScreenWidth/2; i++ {
		c.ViewFix[i] = c.CameraToProjectionPlane / c.TrigTable[c.ScreenWidth/2-1-i].cos
		c.ViewFix[(c.ScreenWidth-1)-i] = c.ViewFix[i]
	}

	c.ZBuffer = make([]float64, c.WorkerWidth*c.ScreenHeight)

}
