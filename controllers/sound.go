// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/ecs"
)

type SoundController struct {
	ecs.BaseController
	*audio.Sound
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &SoundController{} }, 100)
}

func (sc *SoundController) ComponentID() ecs.ComponentID {
	return audio.SoundCID
}

func (sc *SoundController) Methods() ecs.ControllerMethod {
	return ecs.ControllerPrecompute
}

func (sc *SoundController) Target(target ecs.Component, e ecs.Entity) bool {
	sc.Entity = e
	sc.Sound = target.(*audio.Sound)
	return sc.IsActive()
}

func (sc *SoundController) Precompute() {
	if err := sc.Load(); err != nil {
		log.Printf("SoundController.Precompute: %v for %v", err, sc.Source)
	}
}
