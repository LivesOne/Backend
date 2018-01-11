package logger

import (
	"fmt"
	"os"

	"github.com/donnie4w/go-logger/logger"
	"strings"
)

// InitLogger
func InitLogger(dir string,level string) {

	ensureDirExist(dir)

	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	logger.SetRollingDaily(dir, "livesone-backend.log")

	// //指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	// //一般习惯是测试阶段为debug，		 生成环境为info以上
	logger.SetLevel(loggerLevel(level))

}


func loggerLevel(level string)logger.LEVEL{
	level = strings.ToUpper(level)
	switch level {
	case "ALL":
		return logger.ALL
	case "DEBUG":
		return logger.DEBUG
	case "INFO":
		return logger.INFO
	case "WARN":
		return logger.WARN
	case "ERROR":
		return logger.ERROR
	case "FATAL":
		return logger.FATAL
	case "OFF":
		return logger.OFF
	default:
		return logger.ALL

	}
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
	logger.Debug(v)
}

func Error(v ...interface{}) {
	logger.Debug(v)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v)
}

func Info(v ...interface{}) {
	logger.Info(v)
}

func Warn(v ...interface{}) {
	logger.Warn(v)
}
