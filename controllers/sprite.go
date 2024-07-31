// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type SpriteController struct {
	concepts.BaseController
	*materials.Sprite
}

func init() {
	concepts.DbTypes().RegisterController(&SpriteController{})
}

func (sc *SpriteController) ComponentIndex() int {
	return materials.SpriteComponentIndex
}

// Should run before everything
func (a *SpriteController) Priority() int {
	return 40
}

func (a *SpriteController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (a *SpriteController) Target(target concepts.Attachable) bool {
	a.Sprite = target.(*materials.Sprite)
	return a.Sprite.IsActive()
}

func (a *SpriteController) Always() {
	a.Sprite.Image.Refresh(a.Sprite.DB, materials.ImageComponentIndex)
}
