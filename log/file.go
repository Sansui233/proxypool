package log

import (
	"os"
	"path/filepath"
)

var (
	logDir  = "/var/log/proxypool"
	logFile = filepath.Join(logDir, "run.log")
)

func init() {
	//ok := initDir(logDir)
	//fPath := filepath.Join(logDir, logFile)
	//if ok {
	//	if f := initFile(fPath); f != nil {
	//		if err := f.Close(); err != nil {
	//			Infoln("init log file in %s", fPath)
	//		}
	//	}
	//}
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
	// TODO detect old log files and compress
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		Errorln("get log file error: %s", err.Error())
	}
	return f
}
