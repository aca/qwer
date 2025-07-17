package main

import (
	"github.com/aca/qwer/log"
)

func main() {
	log.Info("System starting up")
	log.Warn("High memory usage detected", "memory_usage", "85%")
	log.Error("Database connection failed", "retries", 3)
	log.Emergency("CRITICAL: System failure detected - immediate action required!", 
		"system", "payment_processor",
		"error_code", "SYS_001",
		"affected_users", 15000)
}
