package logger

import (
	"fmt"
	"github.com/alecthomas/log4go"
	"os"
	"path/filepath"
	"github.com/google/uuid"
)

// InitLogger
func InitLogger(dir string, cfgName string) {

	basePath := filepath.Join(dir, cfgName)
	//fmt.Println(filepath.Join(dir, "logs"))
	//ensureDirExist(filepath.Join(dir, "logs"))

	log4go.LoadConfiguration(basePath)

	Info("init logger config path ", basePath)

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

func Error(v ...interface{}) {
	log4go.Error(v)
}

func Info(v ...interface{}) {
	log4go.Info(v)
}

func Warn(v ...interface{}) {
	log4go.Warn(v)
}


type LvtLogger struct {
	LogId string
	infos []interface{}
	LogNow bool
}

func NewLvtLogger(logNow bool)*LvtLogger{
	l := new(LvtLogger)
	l.infos = make([]interface{},0)
	l.LogNow = logNow
	l.LogId = uuid.New().String()
	return l
}

func (l *LvtLogger)Debug(v ...interface{}) {
	if l.LogNow {
		Debug(l.LogId,v)
	}
	l.infos = append(l.infos,v...)
}

func (l *LvtLogger)Info(v ...interface{}) {
	if l.LogNow {
		Info(l.LogId,v)
	}
	l.infos = append(l.infos,v...)
	fmt.Println(l.infos)
}

func (l *LvtLogger)Warn(v ...interface{}) {
	Warn(l.LogId,v)
}

func (l *LvtLogger)Error(v ...interface{}) {
	Error(l.LogId,v)
}

func (l *LvtLogger)InfoAll() {
	logInfo := []interface{}{l.LogId}
	logInfo = append(logInfo,l.infos...)
	Info(logInfo)
}