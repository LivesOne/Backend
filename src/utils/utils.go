package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ReadJSONFile reads a JSON format file into v
func ReadJSONFile(filename string, v interface{}) error {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	return err
}

// GetAppBaseDir get the execute file's absolute path
func GetAppBaseDir() string {
	file, _ := exec.LookPath(os.Args[0])
	fpath, _ := filepath.Abs(file)
	name := filepath.Base(file)
	appDir := strings.TrimRight(fpath, name)
	//fmt.Println("[BaseDir]", appDir)
	return appDir
}

//SetupLogFile 设置日志输出文件
// func SetupLogFile(fileName string) {
// 	appDir := GetAppBaseDir()
// 	fmt.Println("appDir >> ", appDir, "   fileName >> ", fileName)
// 	logFile, err := os.OpenFile(appDir+fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend|0666)
// 	fmt.Println("logFile >> ", logFile)
// 	//fmt.Println("error >> ", err.Error())
// 	if err == nil {
// 		log.SetOutput(logFile)
// 		// log.Println("\n\n\n")
// 		log.SetFlags(log.Flags() | log.Lshortfile)
// 	}
// }
