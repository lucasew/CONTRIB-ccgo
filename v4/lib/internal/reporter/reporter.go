package reporter

import (
	"log"
)

// ReportError funnels all unexpected errors through a centralized reporting function.
func ReportError(msg string, err error) {
	if err != nil {
		log.Printf("ERROR: %s: %v", msg, err)
	}
}
