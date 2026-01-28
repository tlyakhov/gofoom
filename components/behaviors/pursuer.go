// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
	"github.com/srwiley/gheap"
)

type Breadcrumb struct {
	Pos         concepts.Vector3
	TargetTime  int64
	CreatedTime int64
}

type Candidate struct {
	concepts.Ray
	Weight float64
	Count  int
}

type PursuerEnemy struct {
	Entity      ecs.Entity
	Body        *core.Body
	Pos         *concepts.Vector3
	Delta       *concepts.Vector2
	Dist        float64
	InView      bool
	Visited     bool
	Breadcrumbs gheap.MinMaxHeap[int64, Breadcrumb]
}

type Pursuer struct {
	ecs.Attached     `editable:"^"`
	StrafeDistance   float64 `editable:"Strafe Distance"`
	ChaseSpeed       float64 `editable:"Chase Speed"`
	AlwaysFaceTarget bool    `editable:"Always Face Target?"`
	FOV              float64 `editable:"FOV"`
	FireDelay        float64 `editable:"Fire Delay"` //ms

	SoundsTargetSeen ecs.EntityTable `editable:"Sounds (target seen)"`
	SoundsTargetLost ecs.EntityTable `editable:"Sounds (target lost)"`
	SoundsIdle       ecs.EntityTable `editable:"Sounds (idle)"`

	// Internal state (TODO: maybe move into separate component like ActorState?)
	Candidates          []*Candidate
	Enemies             map[ecs.Entity]*PursuerEnemy
	ClockwisePreference bool
	ClockwiseSwitchTime int64
	NextFireTime        int64
	NextIdleBark        int64
}

func (p *Pursuer) String() string {
	return "Pursuer"
}

func (p *Pursuer) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.Enemies = make(map[ecs.Entity]*PursuerEnemy)
	p.ClockwisePreference = rand.UintN(2) == 0
	p.StrafeDistance = 100
	p.ChaseSpeed = 1000
	p.AlwaysFaceTarget = true
	p.FOV = 165
	p.FireDelay = 1000
	p.SoundsTargetSeen = ecs.EntityTable{}
	p.SoundsTargetLost = ecs.EntityTable{}
	p.SoundsIdle = ecs.EntityTable{}
	p.Candidates = make([]*Candidate, 16)
	for i := range p.Candidates {
		p.Candidates[i] = &Candidate{}
	}

	if data == nil {
		return
	}

	if v, ok := data["StrafeDistance"]; ok {
		p.StrafeDistance = cast.ToFloat64(v)
	}
	if v, ok := data["ChaseSpeed"]; ok {
		p.ChaseSpeed = cast.ToFloat64(v)
	}
	if v, ok := data["AlwaysFaceTarget"]; ok {
		p.AlwaysFaceTarget = cast.ToBool(v)
	}
	if v, ok := data["FOV"]; ok {
		p.FOV = cast.ToFloat64(v)
	}
	if v, ok := data["FireDelay"]; ok {
		p.FireDelay = cast.ToFloat64(v)
	}
	if v, ok := data["SoundsTargetSeen"]; ok {
		p.SoundsTargetSeen = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["SoundsTargetLost"]; ok {
		p.SoundsTargetLost = ecs.ParseEntityTable(v, true)
	}
	if v, ok := data["SoundsIdle"]; ok {
		p.SoundsIdle = ecs.ParseEntityTable(v, true)
	}
}

func (p *Pursuer) Serialize() map[string]any {
	result := p.Attached.Serialize()

	result["StrafeDistance"] = p.StrafeDistance
	result["ChaseSpeed"] = p.ChaseSpeed
	if !p.AlwaysFaceTarget {
		result["AlwaysFaceTarget"] = p.AlwaysFaceTarget
	}
	if p.FOV != 165 {
		result["FOV"] = p.FOV
	}
	if p.FireDelay != 1000 {
		result["FireDelay"] = p.FireDelay
	}
	if len(p.SoundsTargetSeen) > 0 {
		result["SoundsTargetSeen"] = p.SoundsTargetSeen.Serialize()
	}
	if len(p.SoundsTargetLost) > 0 {
		result["SoundsTargetLost"] = p.SoundsTargetLost.Serialize()
	}
	if len(p.SoundsIdle) > 0 {
		result["SoundsIdle"] = p.SoundsIdle.Serialize()
	}
	return result
}

func (p *Pursuer) BestCandidate() *Candidate {
	var best *Candidate
	bestWeight := math.Inf(-1)
	for _, c := range p.Candidates {
		if c.Count == 0 {
			continue
		}
		w := c.Weight /// float64(c.Count)
		if w > bestWeight {
			bestWeight = w
			best = c
		}
	}
	return best
}
