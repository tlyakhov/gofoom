// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package resources

import "fyne.io/fyne/v2/theme"

//go:generate fyne bundle --pkg resources -o bundled.go ../../data/resources/icon-add-entity.svg
var ResourceIconAddEntity = theme.NewThemedResource(resourceIconAddEntitySvg)

//go:generate fyne bundle -o bundled.go -append ../../data/resources/icon-add-sector.svg
var ResourceIconAddSector = theme.NewThemedResource(resourceIconAddSectorSvg)

//go:generate fyne bundle -o bundled.go -append ../../data/resources/icon-split-sector.svg
var ResourceIconSplitSector = theme.NewThemedResource(resourceIconSplitSectorSvg)

//go:generate fyne bundle -o bundled.go -append ../../data/resources/icon-split-segment.svg
var ResourceIconSplitSegment = theme.NewThemedResource(resourceIconSplitSegmentSvg)
