// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"fmt"
	"strings"
)

type Flaggable interface {
	~int | ~int16 | ~int32 | ~int64 | ~uint16 | ~uint32 | ~uint64
	fmt.Stringer
}

func ParseFlags[EnumType Flaggable](serialized string, initializer func(string) (EnumType, error)) EnumType {
	parsedFlags := strings.Split(serialized, "|")
	result := EnumType(0)
	for _, parsedFlag := range parsedFlags {
		if f, err := initializer(parsedFlag); err == nil {
			result |= f
		}
	}
	return result
}

func SerializeFlags[EnumType Flaggable](flags EnumType, values []EnumType) string {
	result := ""
	for _, f := range values {
		if flags&f == 0 {
			continue
		}
		if len(result) > 0 {
			result += "|"
		}
		result += f.String()
	}
	return result
}
