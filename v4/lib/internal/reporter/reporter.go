// Copyright 2024 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reporter

import (
	"fmt"
	"os"
)

// ReportError centralizes error reporting across the application.
// It ensures that unexpected errors are handled uniformly.
func ReportError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", msg, err)
	}
}
