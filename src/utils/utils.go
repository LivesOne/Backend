package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"servlets/constants"
	"strconv"
	"strings"
	"time"
	"math"
	"utils/logger"
)

const (
	Second int64 = 1000
	Minute       = 60 * Second
	Hour         = 60 * Minute
	Day          = 24 * Hour
	TwoDay       = 2 * Day

	DayDuration    = 24 * time.Hour
	TwoDayDuration = 2 * DayDuration
	CONV_LVT       = 10000 * 10000
)

// ReadJSONFile reads a JSON format file into v
func ReadJSONFile(filename string, v interface{}) error {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	//logger.Info("read json ",string(data))
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

func Str2Int(str string) int {
	tmp, _ := strconv.Atoi(str)
	return tmp
}

func Int2Str(i int) string {
	return strconv.Itoa(i)
}

func Int642Str(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Str2Int64(str string) int64 {
	tmp, _ := strconv.ParseInt(str, 10, 64)
	return tmp
}

func IsValidEmailAddr(email string) bool {
	ret, _ := regexp.MatchString("^.+@.+$", email)
	return ret
}

func IsDigit(str string) bool {
	ret, _ := regexp.MatchString("^[0-9]*$", str)
	return ret
}

func GetTimestamp13() int64 {
	return GetTimestamp13ByTime(time.Now())
}

func GetTimestamp10() int64 {
	return GetTimestamp10ByTime(time.Now())
}

func GetTimestamp10ByTime(t time.Time) int64 {
	return t.Unix()
}

func GetTimestamp13ByTime(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

// 按 UTC 时间，判断 cur 是不是 last 的第二天
func IsNextDay(last, cur int64) bool {
	lastDate := GetDayStart(last)
	curDate := GetDayStart(cur)
	duration := curDate.Sub(lastDate)
	return duration < DayDuration
}

func IsToday(last, cur int64) bool {
	lastDate := GetDayStart(last)
	curDate := GetDayStart(cur)
	duration := curDate.Sub(lastDate)
	return duration == time.Duration(0)
}

// 获取时间戳 UTC 时间的当日凌晨时间
func GetDayStart(timestamp int64) time.Time {
	timeUtc := Timestamp13ToDate(timestamp)
	year, month, day := timeUtc.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func Timestamp13ToDate(timestamp int64) time.Time {
	second := timestamp / 1000
	nanosecond := timestamp % 1000 * 1000000
	timeLocal := time.Unix(second, nanosecond)
	timeUtc := timeLocal.UTC()
	return timeUtc
}

func LVTintToFloatStr(lvt int64) string {
	return strconv.FormatFloat((float64(lvt) / CONV_LVT), 'f', 8, 64)
}

func FloatStrToLVTint(lvt string) int64 {
	return int64(Str2Float64(lvt) * CONV_LVT)
}

func LVTintToNamorInt(lvt int64)int{
	return int(lvt/CONV_LVT)
}

func NamorFloatToLVTint(nlvt float64)int64{
	return int64(nlvt*CONV_LVT)
}

func Str2Float64(str string) float64 {
	fs, _ := strconv.ParseFloat(str, 64)
	return fs
}

func TXIDToTimeStamp13(txid int64) int64 {
	const timebase int64 = 1514764800000
	tagTs := txid>>22 + constants.BASE_TIMESTAMP
	return tagTs
}

func GetFormatDateNow() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

func GetFormatDateNow14() string {
	return time.Now().UTC().Format("20060102150405")
}

func TimestampToTxid(ts, iv int64) int64 {
	delta := ts - iv - constants.BASE_TIMESTAMP
	tstx := (delta << 22) & 0x7FFFFFFFFFC00000
	return tstx
}


func Round(f float64)int{
	return int(math.Floor(f+0.5))
}


func SignValid(aeskey, signature string, timestamp int64) bool {

	// signature := handler.header.Signature

	if len(signature) < 1 {
		return false
	}

	tmp := aeskey + strconv.FormatInt(timestamp, 10)
	hash := Sha256(tmp)

	res := signature == hash

	if res {
		logger.Info("verify header signature successful", signature, string(hash[:]))
	} else {
		logger.Info("verify header signature failed:", signature, string(hash[:]))
	}

	return res
}


func GetTs13(ts int64)int64{
	if ts > 1000000000 && ts < 2000000000 {
		return ts *1000
	}
	return ts
}



func DecodeSecret(secret, key, iv string, secretPtr interface{}) error {
	if len(secret) == 0 {
		return nil
	}
	logger.Debug("secret ", secret)
	secJson, err := AesDecrypt(secret, key, iv)
	if err != nil {
		logger.Error("aes decode error ", err.Error())
		return err
	}
	logger.Debug("base64 and aes decode secret ", secJson)
	err = json.Unmarshal([]byte(secJson), secretPtr)
	if err != nil {
		logger.Error("json Unmarshal error ", err.Error())
		return err
	}
	return nil

}
