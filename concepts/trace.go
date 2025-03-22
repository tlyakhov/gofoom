// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// Use the below like: `defer ExecutionDuration(ExecutionTrack("blah"))`
func ExecutionTrack(msg string) (string, time.Time) {
	return msg, time.Now()
}

func ExecutionDuration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}

func StackTrace() string {
	var sb strings.Builder
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return sb.String()
}
