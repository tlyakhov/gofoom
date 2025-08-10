// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"fmt"
	"strconv"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=ProximityStatus -json
type ProximityStatus int

const (
	ProximityIdle ProximityStatus = iota
	ProximityFiring
	ProximityWaiting
)

type ProximityState struct {
	ecs.Attached

	LastFired  int64
	Status     ProximityStatus
	PrevStatus ProximityStatus
	Source     ecs.Entity
	Target     ecs.Entity
	Flags      ProximityFlags
}

func (p *ProximityState) String() string {
	return fmt.Sprintf("ProximityState (%v is close to %v, status %v)", p.Target.String(), p.Source.String(), p.Status.String())
}

func (p *ProximityState) Construct(data map[string]any) {
	p.Attached.Construct(data)
	p.Attached.Flags |= ecs.ComponentInternal

	if data == nil {
		return
	}

	if v, ok := data["Source"]; ok {
		p.Source, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["Target"]; ok {
		p.Target, _ = ecs.ParseEntity(v.(string))
	}
	if v, ok := data["LastFired"]; ok {
		p.LastFired = cast.ToInt64(v)
	}
	if v, ok := data["Status"]; ok {
		p.Status, _ = ProximityStatusString(v.(string))
	}
	if v, ok := data["Flags"]; ok {
		p.Flags, _ = ProximityFlagsString(v.(string))
	}
}

func (p *ProximityState) Serialize() map[string]any {
	result := p.Attached.Serialize()
	result["Source"] = p.Source.Serialize()
	result["Target"] = p.Target.Serialize()
	result["LastFired"] = strconv.FormatInt(p.LastFired, 10)
	result["Status"] = p.Status.String()
	result["Flags"] = p.Flags.String()

	return result
}
