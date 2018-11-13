package utils

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"servlets/constants"
	"strconv"
	"strings"
	"time"
	"utils/logger"
	"reflect"
)

const (
	Second int64 = 1000
	Minute       = 60 * Second
	Hour         = 60 * Minute
	Day          = 24 * Hour
	TwoDay       = 2 * Day

	DayDuration         = 24 * time.Hour
	TwoDayDuration      = 2 * DayDuration
	CONV_LVT            = 1e8
	CONV_EOS            = 1e4
	DB_CONV_CHAIN_VALUE = 1e10

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

func CoinsInt2FloatStr(coins, coinsDecimal int64) string {
	return strconv.FormatFloat(float64(coins) / float64(coinsDecimal), 'f', 8, 64)
}

func FloatStr2CoinsInt (coins string, coinsDecimal int64) int64 {
	return int64(Str2Float64(coins) * float64(coinsDecimal))
}

func Float2CoinsInt (coins float64, coinsDecimal int64) int64 {
	return int64(coins * float64(coinsDecimal))
}

func LVTintToFloatStr(lvt int64) string {
	d2 := decimal.New(lvt, 0).Div(decimal.NewFromFloat(CONV_LVT))
	return d2.StringFixed(8)
}

func EOSintToFloatStr(lvt int64) string {
	d2 := decimal.New(lvt, 0).Div(decimal.NewFromFloat(CONV_EOS))
	return d2.StringFixed(4)
}

func FloatStrToLVTint(lvt string) int64 {

	d2, err := decimal.NewFromString(lvt)
	if err != nil {
		logger.Error("decimal conv folat error", err.Error())
		return 0
	}
	d3 := d2.Mul(decimal.NewFromFloat(CONV_LVT))

	return d3.IntPart()
}

func FloatStrToEOSint(eos string) int64 {

	d2, err := decimal.NewFromString(eos)
	if err != nil {
		logger.Error("decimal conv folat error", err.Error())
		return 0
	}
	d3 := d2.Mul(decimal.NewFromFloat(CONV_EOS))

	return d3.IntPart()
}

func LVTintToNamorInt(lvt int64) int {
	d := decimal.New(lvt, 0).Div(decimal.NewFromFloat(CONV_LVT))
	return int(d.IntPart())
}

func NamorFloatToLVTint(nlvt float64) int64 {
	d := decimal.NewFromFloat(nlvt).Mul(decimal.NewFromFloat(CONV_LVT))
	return d.IntPart()
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

func Round(f float64) int {
	return int(math.Floor(f + 0.5))
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

func GetTs13(ts int64) int64 {
	if ts > 1000000000 && ts < 2000000000 {
		return ts * 1000
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
	logger.Debug("decode secret",ToJSON(secretPtr))
	return nil

}

func ConvDBValueToChainStr(lvt int64) string {
	dbValue := big.NewInt(lvt)
	withdrawal := big.NewInt(0).Mul(dbValue, big.NewInt(DB_CONV_CHAIN_VALUE))
	return withdrawal.Text(10)
}

func ConvChainStrToDBValue(chainStr string) int64 {
	chainValue, ok := new(big.Int).SetString(chainStr, 10)
	if !ok {
		return 0
	}
	dbValue := big.NewInt(0).Div(chainValue, big.NewInt(DB_CONV_CHAIN_VALUE))
	return dbValue.Int64()
}



func GetLockHashrate(lvtcScale,monnth int, value string) int {
	//锁仓数额	B	[用户自定义填充]，锁仓额为1000LVT的倍数
	b := Str2Float64(value)
	//锁仓期间	T	用户选择：1个月、3个月、6个月、12个月，24个月
	t := float64(monnth)

	//算力系数 a=0.2 计算算力为整数，a=0.2 扩大100倍 a := 20
	a := float64(20)
	//锁仓算力	S	S=lvtcScale*B/100000*T*a*100%（a=0.2）
	s := float64(lvtcScale) * b / 100000 * t * a

	//Mmax=500%，大于500%取500%
	//四舍五入后数值大于500 取500
	if lvtcScale > 0 {
		if re := Round(s); re <= constants.ASSET_LOCK_MAX_VALUE {
			return re
		}
	}
	return constants.ASSET_LOCK_MAX_VALUE
}



func StructConvMap(p interface{}) map[string]interface{} {
	v,t := GetStructValueAndType(p)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Tag.Get("json")
		if name == "-" {
			continue
		}
		if len(name) == 0 {
			name = f.Name
		}

		value :=  v.Field(i).Interface()
		valueStr := convStructField(value)
		if nss := strings.Split(name,",");len(nss)>1{
			name = nss[0]
			if nss[1] == "omitempty" {
				if len(valueStr) == 0 {
					continue
				}
			}
		}
		data[name] = value
	}
	return data
}

func GetStructValueAndType(p interface{})(reflect.Value,reflect.Type){
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	return v,v.Type()
}


func convStructField(p interface{}) string {
	switch p.(type) {
	case int:
		s := p.(int)
		if s != 0 {
			return Int2Str(s)
		}
	case int64:
		s := p.(int64)
		if s != 0 {
			return Int642Str(s)
		}
	case float64:
		s := p.(float64)
		if s != 0 {
			return  strconv.FormatFloat(s,'f', 8, 64)
		}
	case string:
		return p.(string)
	}
	return ""
}


func GetTomorrowStartTs10()int64{
	k := time.Now().UTC()
	d, _ := time.ParseDuration("+24h")
	k = k.Add(d)
	return GetTimestamp10ByTime(GetDayStart(GetTimestamp13ByTime(k)))
}

func Scientific2Str(srcStr string) string {
	var new float64
	fmt.Sscanf(srcStr, "%e", &new)
	return strconv.FormatFloat(new,'f',-1,64)
}

func Float642Str(value float64) string {
	return strconv.FormatFloat(value,'f',-1,64)
}
