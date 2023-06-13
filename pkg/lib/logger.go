package lib

import (
	"os"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

var (
	log     *logrus.Entry
	logFile *os.File
)

func InitLogger(serviceName string, logFileLocation string) (*logrus.Entry, *os.File) {
	// logFileLocation := fmt.Sprintf("/var/log/service_logs/%s.log", serviceName)
	// if os.Getenv("MODE") == "local" {
	// 	mydir, _ := os.Getwd()
	// 	logFileLocation = fmt.Sprintf("%s/%s.log", mydir, serviceName)
	// }

	logFile, err := os.OpenFile(logFileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("Failed to open logfile %s: %v", serviceName, err)
	}

	env := os.Getenv("Environment")

	if env == "Production" {
		// Only log the Info level severity or above.
		logrus.SetLevel(logrus.WarnLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set Logrus output to the log file
	logrus.SetOutput(logFile)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Disable to prevent 20%-40% overhead
	logrus.SetReportCaller(true)

	log = logrus.WithFields(logrus.Fields{
		"service": serviceName,
	})

	return log, logFile
}

// New function to flush the logs
func FlushLogs(logFile *os.File) {
	logFile.Sync()
}

func HandlePanicAndFlushLogs(log *logrus.Entry, logFile *os.File) {
	if r := recover(); r != nil {
		stackTrace := string(debug.Stack())
		log.WithField("stackTrace", stackTrace).Errorf("Panic occurred: %v", r)
		FlushLogs(logFile) // Flush the logs before exiting
		panic(r)
	}
}
