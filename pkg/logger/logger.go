package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new structured logger with the specified level
func NewLogger(level string) *logrus.Logger {
	log := logrus.New()
	
	// Set log format
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	
	// Set output to stdout
	log.SetOutput(os.Stdout)
	
	// Parse and set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	
	return log
}