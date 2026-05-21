package reporter

import (
	"log"
)

// ReportError centralizes error reporting for unexpected issues.
func ReportError(err error) {
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
}
