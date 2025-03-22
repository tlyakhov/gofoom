// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import "unicode"

// Adapted from https://stackoverflow.com/questions/59955085/how-can-i-elliptically-truncate-text-in-golang
func TruncateString(text string, limit int) string {
	lastWordBoundaryIndex := limit
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			lastWordBoundaryIndex = i
		}
		len++
		if len > limit {
			return text[:lastWordBoundaryIndex] + "..."
		}
	}
	// If here, string is shorter or equal to limit
	return text
}
