package logger

import (
	"fmt"
	"os"
	"github.com/alecthomas/log4go"
	"path/filepath"
)

// InitLogger
func InitLogger(cfgPath string,dir string) {

	if len(cfgPath) == 0 {
		cfgPath = filepath.Join(dir, "../config/l4g.xml")
	}


	//fmt.Println(filepath.Join(dir, "logs"))
	//ensureDirExist(filepath.Join(dir, "logs"))

	log4go.LoadConfiguration(cfgPath)

	Info("init logger config path ",cfgPath)

}




func ensureDirExist(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(dir, os.ModeDir|0777)
			return
		} else {
			fmt.Println("create logs dir error:", err)
		}
	}

}

func Debug(v ...interface{}) {
	log4go.Debug(v)
}

func Error(v  ...interface{}) {
	log4go.Error(v)
}


func Info(v ...interface{}) {
	log4go.Info(v)
}

func Warn(v ...interface{}) {
	log4go.Warn(v)
}
