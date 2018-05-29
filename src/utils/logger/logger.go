package logger

import (
	"fmt"
	"github.com/alecthomas/log4go"
	"os"
	"path/filepath"
	"github.com/google/uuid"
	"runtime"
	"strings"
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
	LogNow bool
	infos []interface{}

}

func NewLvtLogger(logNow bool,info ...interface{})*LvtLogger{
	l := new(LvtLogger)
	l.infos = make([]interface{},0)
	l.LogNow = logNow
	l.LogId = uuid.New().String()
	if _, file, _, ok := runtime.Caller(1);ok{
		fs := strings.Split(file,"/")
		l.infos = append(l.infos,"file : "+fs[len(fs)-1])
	}
	if len(info) >0 {
		l.Info(info...)
	}
	return l
}

func (l *LvtLogger)Debug(v ...interface{}) {
	if l.LogNow {
		i := append([]interface{}{l.LogId},v...)
		Debug(i...)
	}
	l.infos = append(l.infos,v...)
}

func (l *LvtLogger)Info(v ...interface{}) {
	if l.LogNow {
		i := append([]interface{}{l.LogId},v...)
		Info(i...)
	}
	l.infos = append(l.infos,v...)
}

func (l *LvtLogger)Warn(v ...interface{}) {
	if l.LogNow {
		i := append([]interface{}{l.LogId},v...)
		Warn(i...)
	}
	l.infos = append(l.infos,v...)
}

func (l *LvtLogger)Error(v ...interface{}) {
	if l.LogNow {
		i := append([]interface{}{l.LogId},v...)
		Error(i...)
	}
	l.infos = append(l.infos,v...)
}

func (l *LvtLogger)InfoAll() {
	Info(l.infos...)
}