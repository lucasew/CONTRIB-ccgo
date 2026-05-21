// Copyright 2025 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package reporter provides a centralized error-reporting function.
package reporter

import (
	"log"
	"runtime/debug"
)

// ReportError is the centralized error reporting function for the project.
// All code paths that handle unexpected errors MUST funnel through this function.
func ReportError(err error) {
	if err == nil {
		return
	}

	// Since there is no Sentry set up, we log the error with its stack trace.
	log.Printf("ERROR: %v\nStack Trace:\n%s", err, debug.Stack())
}
