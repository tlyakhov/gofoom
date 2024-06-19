// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

//go:build tools

package concepts

import (
	_ "github.com/dmarkham/enumer"
	_ "github.com/klauspost/cpuid/v2"
	_ "github.com/mmcloughlin/avo"
	_ "golang.org/x/tools/cmd/stringer"
)
