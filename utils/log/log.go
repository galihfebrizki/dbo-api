package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func InitLog(env string, logLevel int) {
	log.New()

	if env != "development" {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.Level(logLevel))

	// Get reporter file and function
	log.SetReportCaller(true)
}
