package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"sync"
)

var (
	level      = INFO
	fileLogger = log.New()
	fileMux    = sync.Mutex{}
)

func init() {
	log.SetFormatter(&prefixed.TextFormatter{
		ForceFormatting: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	fileLogger.SetFormatter(&prefixed.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
		ForceFormatting: true,
	})
	fileLogger.SetLevel(levelMapping[TRACE])
}

func SetLevel(l LogLevel) {
	level = l
	log.SetLevel(levelMapping[level])
}

func Traceln(format string, v ...interface{}) {
	log.Traceln(fmt.Sprintf(format, v...))
	logToFile(TRACE, fmt.Sprintf(format, v...))
}

func Debugln(format string, v ...interface{}) {
	log.Debugln(fmt.Sprintf(format, v...))
	logToFile(DEBUG, fmt.Sprintf(format, v...))
}

func Infoln(format string, v ...interface{}) {
	log.Infoln(fmt.Sprintf(format, v...))
	logToFile(INFO, fmt.Sprintf(format, v...))
}

func Warnln(format string, v ...interface{}) {
	log.Warnln(fmt.Sprintf(format, v...))
	logToFile(WARNING, fmt.Sprintf(format, v...))
}

func Errorln(format string, v ...interface{}) {
	log.Errorln(fmt.Sprintf(format, v...))
	logToFile(ERROR, fmt.Sprintf(format, v...))
}

func logToFile(l LogLevel, data string) {
	if l >= level {
		if logFile != nil {
			fileMux.Lock()
			fileLogger.SetOutput(logFile)
			fileLogger.Logln(levelMapping[l], data)
			fileMux.Unlock()
		}
	}
	if allLogFile != nil {
		fileMux.Lock()
		fileLogger.SetOutput(allLogFile)
		fileLogger.Logln(levelMapping[l], data)
		fileMux.Unlock()
	}
}
