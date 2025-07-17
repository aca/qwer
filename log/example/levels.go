package main

import (
	"github.com/aca/qwer/log"
)

func main() {
	// Demonstrate all log levels from lowest to highest severity
	log.Trace("TRACE: Very detailed diagnostic information")
	log.Debug("DEBUG: Detailed information for debugging")
	log.Info("INFO: General informational messages")
	log.Warn("WARN: Warning messages about potential issues")
	log.Error("ERROR: Error messages for failures")
	log.Emergency("EMERGENCY: Critical failures requiring immediate attention")
}
