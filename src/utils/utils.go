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
	"time"
)


const (
	Second int64 = 1000
	Minute = 60 * Second
	Hour = 60 * Minute
	Day = 24 * Hour
	TwoDay = 2 * Day

	DayDuration  = 24 * time.Hour
	TwoDayDuration = 2 * DayDuration
	CONV_LVT = 10000*10000
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

func Int2Str(i int)string  {
	return strconv.Itoa(i)
}

func Int642Str(i int64)string  {
	return strconv.FormatInt(i,10)
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
		res := string(body)
		logger.Info("SendPost res ---> ", res)
		return res, e2
	}
}
func GetTimestamp13() int64 {
	now := time.Now()
	return now.UnixNano() / 1000000
}

// 按 UTC 时间，判断 cur 是不是 last 的第二天
func IsNextDay(last, cur int64) bool {
	lastDate := GetDayStart(last)
	curDate := GetDayStart(cur)
	duration := curDate.Sub(lastDate)
	if (duration >= DayDuration) && (TwoDayDuration > duration) {
		return true
	} else {
		return false
	}
}

// 获取时间戳 UTC 时间的当日凌晨时间
func GetDayStart(timestamp int64) time.Time {
	timeUtc := Timestamp13ToDate(timestamp)
	year, month, day := timeUtc.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func Timestamp13ToDate(timestamp int64) time.Time {
	second := timestamp/1000
	nanosecond := timestamp % 1000 * 1000000
	timeLocal := time.Unix(second, nanosecond)
	timeUtc := timeLocal.UTC()
	return timeUtc
}

func LVTintToFloatStr(lvt int64)string{
	return strconv.FormatFloat((float64(lvt) / CONV_LVT),'f',8,64)
}


func FloatStrToLVTint(lvt string)int64{
	fs,_ := strconv.ParseFloat(lvt,64)
	return int64(fs*CONV_LVT)
}
