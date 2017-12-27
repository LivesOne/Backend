package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"net/http"
	"utils/logger"
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

func Str2Int(str string) int {
	tmp, _ := strconv.Atoi(str)
	return tmp
}

func Str2Int64(str string) int64 {
	tmp, _ := strconv.ParseInt(str, 10, 64)
	return tmp
}

func IsValidEmailAddr(email string) bool {
	ret, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", email)
	return ret
}


//发起post请求
func Post(url string, params string) (resBody string, e error) {
	logger.Info("SendPost ---> ", url)
	logger.Info("SendPost param ---> ", params)
	resp, e1 := http.Post(url, "application/json", strings.NewReader(params))
	if e1 != nil {
		logger.Error("post error ---> ", e1.Error())
		return "", e1
	} else {
		defer resp.Body.Close()
		body, e2 := ioutil.ReadAll(resp.Body)
		if e2 != nil {
			logger.Error("post error ---> ", e2.Error())
		}
		return string(body), e2
	}
}
