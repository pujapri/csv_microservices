package logger

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Initialize the logger

func InitializeLogger() {
	Log = logrus.New()
	Log.SetOutput(&lumberjack.Logger{
		Filename:   "File.log",
		MaxSize:    10,   // Max size in MB before rotating,  If File.log exceeds 10MB, it will be archived, and a new log file will be created automatically.
		MaxBackups: 3,    // Max number of old log files to keep
		MaxAge:     7,    // Max age in days to keep old log files
		Compress:   true, // Compress old log files
	})
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel) // Log level is set to INFO, so only informational and more critical messages will be logged.
}
