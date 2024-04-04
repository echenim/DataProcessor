package logger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Log is an exported Logger instance to use throughout your application.
var Log *logrus.Logger

func Setup(dirPath string) {
	Log = logrus.New()

	// Set the level of the logger. In production, this might be logrus.InfoLevel.
	Log.SetLevel(logrus.DebugLevel)

	// Set formatter
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	rootDir := dirPath
	// Setup log file
	logFilePath := rootDir + "/logs/app.log"
	setupLogFile(logFilePath)

	// Example of how to log
	Log.Info("Logger setup complete")
}

func setupLogFile(logFilePath string) {
	// Ensure the directory exists
	err := os.MkdirAll(filepath.Dir(logFilePath), 0o755)
	if err != nil {
		Log.WithField("error", err).Fatal("Failed to create log directory")
	}

	// Open or create the log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		Log.WithField("error", err).Fatal("Failed to open log file")
	}

	// Set output to the file
	Log.SetOutput(file)
}

func Error(args ...interface{}) {
	log.Println(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}
