package state

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
)

type trigEntry struct {
	sin, cos float64
}

type Config struct {
	ScreenWidth, ScreenHeight int
	Frame, Counter            int
	MaxViewDist, FOV          float64
	CameraToProjectionPlane   float64
	ViewRadians               []float64
	ViewFix                   []float64
	ZBuffer                   []float64
}

var FloorNormal concepts.Vector3 = concepts.Vector3{0, 0, 1}
var CeilingNormal concepts.Vector3 = concepts.Vector3{0, 0, -1}

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
