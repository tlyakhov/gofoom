package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"

	"github.com/loov/hrtime"
)

// This is based on the "Fix your timestep" blog post here:
// https://gafferongames.com/post/fix_your_timestep/
type Simulation struct {
	SimTime          float64
	RenderTime       float64
	RenderStateBlend float64
	FPS              float64
	PrevTimestamp    int64
	Integrate        func()
	Render           func()
	AllScalars       map[*SimScalar]bool
	AllVector2s      map[*SimVector2]bool
	AllVector3s      map[*SimVector3]bool
}

type Simulated interface {
	Attach(sim *Simulation)
	Detach()
	Sim() *Simulation
}

type SimScalar struct {
	Now            float64
	Prev           float64
	Original       float64 `editable:"Initial Value"`
	Render         float64
	RenderCallback func()
}

type SimVector2 struct {
	Now            concepts.Vector2
	Prev           concepts.Vector2
	Original       concepts.Vector2 `editable:"Initial Value"`
	Render         concepts.Vector2
	RenderCallback func()
}

type SimVector3 struct {
	Now            concepts.Vector3
	Prev           concepts.Vector3
	Original       concepts.Vector3 `editable:"Initial Value"`
	Render         concepts.Vector3
	RenderCallback func()
}

func (s *SimScalar) Reset() {
	s.Prev = s.Original
	s.Now = s.Original
}

func (s *SimScalar) Set(v float64) {
	s.Original = v
	s.Reset()
}

func (s *SimScalar) Attach(sim *Simulation) {
	sim.AllScalars[s] = true
}

func (s *SimScalar) Detach(sim *Simulation) {
	delete(sim.AllScalars, s)
}

func (s *SimVector2) Reset() {
	s.Prev = s.Original
	s.Now = s.Original
}

func (s *SimVector2) Set(x float64, y float64) {
	s.Original[1] = y
	s.Original[0] = x
	s.Reset()
}

func (s *SimVector2) Attach(sim *Simulation) {
	sim.AllVector2s[s] = true
}

func (s *SimVector2) Detach(sim *Simulation) {
	delete(sim.AllVector2s, s)
}

func (v *SimVector2) Serialize() map[string]interface{} {
	return v.Original.Serialize()
}

func (v *SimVector2) Deserialize(data map[string]interface{}) {
	v.Original.Deserialize(data)
	v.Reset()
}

func (s *SimVector3) Reset() {
	s.Prev = s.Original
	s.Now = s.Original
}

func (s *SimVector3) Set(x float64, y float64, z float64) {
	s.Original[2] = z
	s.Original[1] = y
	s.Original[0] = x
	s.Reset()
}

func (s *SimVector3) Attach(sim *Simulation) {
	sim.AllVector3s[s] = true
}

func (s *SimVector3) Detach(sim *Simulation) {
	delete(sim.AllVector3s, s)
}

func (v *SimVector3) Serialize() map[string]interface{} {
	return v.Original.Serialize()
}

func (v *SimVector3) Deserialize(data map[string]interface{}) {
	v.Original.Deserialize(data)
	v.Reset()
}

func NewSimulation() *Simulation {
	return &Simulation{
		PrevTimestamp: hrtime.Now().Milliseconds(),
		AllScalars:    make(map[*SimScalar]bool),
		AllVector2s:   make(map[*SimVector2]bool),
		AllVector3s:   make(map[*SimVector3]bool),
	}
}

func (s *Simulation) Step() {
	newTimestamp := hrtime.Now().Milliseconds()
	frameMillis := float64(newTimestamp - s.PrevTimestamp)
	s.PrevTimestamp = newTimestamp
	if frameMillis != 0 {
		s.FPS = 1000.0 / frameMillis
	}

	if frameMillis > constants.MinMillisPerFrame {
		frameMillis = constants.MinMillisPerFrame
	}

	s.RenderTime += frameMillis

	for s.RenderTime >= constants.TimeStep {
		for v := range s.AllScalars {
			v.Prev = v.Now
		}
		for v := range s.AllVector2s {
			v.Prev = v.Now
		}
		for v := range s.AllVector3s {
			v.Prev = v.Now
		}

		s.Integrate()
		s.RenderTime -= constants.TimeStep
		s.SimTime += constants.TimeStep
	}

	// Update the blended values
	s.RenderStateBlend = s.RenderTime / constants.TimeStep

	for v := range s.AllScalars {
		v.Render = v.Now*s.RenderStateBlend + v.Prev*(1.0-s.RenderStateBlend)
		if v.RenderCallback != nil {
			v.RenderCallback()
		}
	}
	for v := range s.AllVector2s {
		v.Render[1] = v.Now[1]*s.RenderStateBlend + v.Prev[1]*(1.0-s.RenderStateBlend)
		v.Render[0] = v.Now[0]*s.RenderStateBlend + v.Prev[0]*(1.0-s.RenderStateBlend)
		if v.RenderCallback != nil {
			v.RenderCallback()
		}
	}
	for v := range s.AllVector3s {
		v.Render[2] = v.Now[2]*s.RenderStateBlend + v.Prev[2]*(1.0-s.RenderStateBlend)
		v.Render[1] = v.Now[1]*s.RenderStateBlend + v.Prev[1]*(1.0-s.RenderStateBlend)
		v.Render[0] = v.Now[0]*s.RenderStateBlend + v.Prev[0]*(1.0-s.RenderStateBlend)
		if v.RenderCallback != nil {
			v.RenderCallback()
		}
	}

	s.Render()
}
