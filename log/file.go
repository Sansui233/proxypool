package log

import (
	"os"
	"path/filepath"
)

var (
	logDir         = "tmp"
	logFilePath    = filepath.Join(logDir, "run.log")
	allLogFilePath = filepath.Join(logDir, "all.log")
)

var logFile *os.File
var allLogFile *os.File

func init() {
	ok := initDir(logDir)
	if ok {
		logFile = initFile(logFilePath)
		allLogFile = initFile(allLogFilePath)
	}
}

func initDir(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			Errorln("init log dir error: %s", err.Error())
		}
	}
	return true
}

func initFile(path string) *os.File {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		Errorln("get log file error: %s", err.Error())
	}
	return f
}
