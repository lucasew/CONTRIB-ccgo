package reporter

import (
	"log"
)

// ReportError logs the error with its context instead of swallowing it.
func ReportError(context string, err error) {
	if err != nil {
		log.Printf("ERROR [%s]: %v", context, err)
	}
}
