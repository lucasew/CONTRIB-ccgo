package reporter

import (
	"log"
)

// ReportError logs unexpected errors. This is the centralized error reporting function.
func ReportError(context string, err error) {
	if err != nil {
		log.Printf("ERROR: %s: %v\n", context, err)
	}
}
