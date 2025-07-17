package main

import (
	"log/slog"

	"github.com/aca/qwer/log"
)

func main() {
	// Using log directly - it already has TextHandler with AddSource and Trace level
	log.Info("hello info", "version", "1.0.0")
	slog.Debug("hello debug", "version", "1.0.0")
	log.Trace("hello trace", "version", "1.0.0")
}
