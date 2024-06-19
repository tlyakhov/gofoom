// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package resources

import "fyne.io/fyne/v2/theme"

//go:generate fyne bundle --pkg resources -o bundled.go ../../data/icon-entity-add.svg
var ResourceIconEntityAdd = theme.NewThemedResource(resourceIconEntityAddSvg)

//go:generate fyne bundle -o bundled.go -append ../../data/icon-sector-add.svg
var ResourceIconSectorAdd = theme.NewThemedResource(resourceIconSectorAddSvg)
