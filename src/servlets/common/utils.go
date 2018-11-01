package common

import (
	"crypto/rand"
	"encoding/json"
	"hash/crc32"
	"io"
	"net/http"
	"servlets/constants"
	"strconv"
	"utils"
	"utils/logger"

	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"strings"
)

// FlushJSONData2Client flush json data to http Client
func FlushJSONData2Client(data interface{}, writer http.ResponseWriter) {

	if (data == nil) || (writer == nil) {
		logger.Info("FlushJSONData2Client: internel error, data or writer is nil pointer")
		return
	}

	writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	writer.WriteHeader(http.StatusOK)

	toClient, err := json.Marshal(data)
	if err == nil {
		writer.Write(toClient)
		logger.Info("FlushJsonData2Clinet data success:\n", string(toClient))
	} else {
		logger.Info("FlushJsonData2Clinet data error: ", err.Error())
	}
}

func ParseHttpHeaderParams(request *http.Request) *HeaderParams {

	params := &HeaderParams{
		TokenHash: request.Header.Get("Token-Hash"),
		Signature: request.Header.Get("Signature"),
		Timestamp: -1, // if parse timestamp error, keep it as default: -1
	}

	time := request.Header.Get("Timestamp")
	if len(time) > 0 {
		ts, err := strconv.ParseInt(time, 10, 64)
		if err == nil {
			params.Timestamp = ts
		} else {
			logger.Info("ParseHttpHeaderParams: parse timestamp error: ", err)
		}
	}

	logger.Info("ParseHttpHeaderParams: received http header: ", utils.ToJSONIndent(params))

	return params
}

// ParseHttpBodyParams parse the http request body params
func ParseHttpBodyParams(request *http.Request, body interface{}) bool {

	logger.Info("ParseHttpBodyParams: request.ContentLength: ", request.ContentLength)
	if request.ContentLength < 1 {
		return true
	}

	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Info("ParseHttpBodyParams: read http body error : ", err)
		return false
	}

	logger.Info("ParseHttpBodyParams: read http body: ", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, body)
	if err != nil {
		logger.Info("ParseHttpBodyParams: parse body param error: ", err)
		return false
	}
	logger.Info("ParseHttpBodyParams: read http request body success:", utils.ToJSON(body))

	return true
}

func GenerateUID() string {

	s := "0123456789"
	box := []byte(s)

	len := 9
	var uid string

	for true {
		uid = "1"
		i := 0
		for {
			if i > len-3 {
				break
			}

			r := make([]byte, 16)
			rand.Read(r)
			index := int(r[0]) % 10
			uid += string(box[index])

			i++
		}

		ieee := crc32.NewIEEE()
		io.WriteString(ieee, uid)
		sum := ieee.Sum32()

		crc := int(sum) % 10
		uid += string(box[crc])

		if IsGoodUID(uid) {
			logger.Info("Good UID: ", uid)
			continue
		}
		break
	}

	return uid
}

func IsGoodUID(uid string) bool {

	// check same or increase or decrease number
	var last_number byte
	numbers := []byte(uid)
	count_same := 0
	count_incr := 0
	count_decr := 0
	for _, number := range numbers {
		if number == last_number {
			count_same++
			count_incr = 1
			count_decr = 1
		} else if number == last_number+1 {
			count_incr++
			count_same = 1
			count_decr = 1
		} else if number == last_number-1 {
			count_decr++
			count_same = 1
			count_incr = 1
		} else {
			count_same = 1
			count_incr = 1
			count_decr = 1
		}

		last_number = number

		if count_same >= 4 || count_incr >= 4 || count_decr >= 4 {
			return true
		}
	}

	// check fixed sub string
	var fixed_strings = []string{"666", "686", "688", "866", "868", "886", "888", "999", "000",
		"1688", "2688", "2008", "2088", "5188", "5201314", "5211314", "10010", "10086"}

	for _, v := range fixed_strings {
		if strings.Contains(uid, v) {
			return true
		}
	}

	return false
}

// GenerateTxID generate a new transaction ID
func GenerateTxID() int64 {
	rid, err := getAutoIncrID("id_tx")
	if err != constants.ERR_INT_OK {
		return -1
	}

	// rid ONLY live in lower 22 bits
	rid = rid & 0x00000000003FFFFF

	//const timebase int64 =
	delta := utils.GetTimestamp13() - constants.BASE_TIMESTAMP // Jan 1, 2018, 00:00:00

	txid := (delta << 22) & 0x7FFFFFFFFFC00000 // move left 22 bit
	txid += int64(rid)

	// fmt.Println("id from redis:", rid, " txid:", txid)

	return txid
}

// getTxID gets the INCR tx ID from the redis
// put this function in the redis_db.go file && call it from the GenerateTxID()
//        causes "import cycle" error
func getAutoIncrID(key string) (int64, int) {
	conn := GetRedisConn()
	if conn == nil {
		return -1, constants.ERR_INT_TK_DB
	}
	defer conn.Close()

	// idx, err := redis.Int(conn.Do("INCR", key))
	reply, err := conn.Do("INCR", key)
	if err != nil {
		return -1, constants.ERR_INT_TK_DB
	} else if reply == nil {
		return -1, constants.ERR_INT_TK_NOTEXISTS
	} else {
		idx, _ := redis.Int64(reply, nil)
		return idx, constants.ERR_INT_OK
	}
}

func GenerateTradeNo(type_id, subtype_id int) string {
	datetime_str := utils.GetFormatDateNow14()
	ver := 1
	cluster_id := 1

	rid, err := getAutoIncrID("id_trade")
	if err != constants.ERR_INT_OK {
		return ""
	}

	rid = rid % 10000

	trade_no := fmt.Sprintf("%s%02d%03d%03d%02d%04d", datetime_str, ver, type_id, subtype_id, cluster_id, rid)
	return trade_no
}

func GenerateMemoFromUID(uid string) string {
	l := len(uid)
	if l != 9 || strings.Index(uid, "1") != 0 {
		return ""
	}

	src := uid[1:]
	seed_pos := l - 4
	seed, _ := strconv.Atoi(string(src[seed_pos]))
	offset := seed % 5 + 2

	data := make([]int, 0, 8)
	for i, v := range src {
		if i == seed_pos {
			continue
		}

		n, _ := strconv.Atoi(string(v))
		data = append(data, (n + 10 - offset) % 10)
	}

	idx := 0
	ret := make([]int, 0)
	sub_data := make([]int, offset)
	for i, v := range data {
		sub_data[idx] = v
		idx++
		if idx == offset || i == len(data) - 1 {
			for j := 0; j < idx/2; j++{
				t := sub_data[j]
				sub_data[j] = sub_data[idx - j - 1]
				sub_data[idx -j - 1] = t
			}

			for i, v := range sub_data {
				if i < idx {
					ret = append(ret, v)
				}
			}
			idx = 0
		}
	}

	// insert seed
	memo := make([]string, 0)
	for i, v := range ret {
		if i == 2 {
			memo = append(memo, strconv.Itoa(seed))
		}
		memo = append(memo, strconv.Itoa(v))
	}

	return strings.Join(memo, "")
}

func ParseUIDFromMemo(memo string) string {
	l := len(memo)
	if l != 8 {
		return ""
	}

	src := memo[0:]

	seed_pos := 2
	seed, _ := strconv.Atoi(string(src[seed_pos]))
	offset := seed % 5 + 2
	data := make([]int, 0, 8)
	for i, v := range src {
		if i == seed_pos {
			continue
		}

		n, _ := strconv.Atoi(string(v))
		data = append(data, (n + offset) % 10)
	}

	idx := 0
	ret := make([]int, 0)
	sub_data := make([]int, offset)
	for i, v := range data {
		sub_data[idx] = v
		idx++
		if idx == offset || i == len(data) - 1 {
			for j := 0; j < idx/2; j++{
				t := sub_data[j]
				sub_data[j] = sub_data[idx - j - 1]
				sub_data[idx -j - 1] = t
			}

			for i, v := range sub_data {
				if i < idx {
					ret = append(ret, v)
				}
			}
			idx = 0
		}
	}

	// insert seed
	uid := []string {"1"}
	for i, v := range ret {
		if i == 5 {
			uid = append(uid, strconv.Itoa(seed))
		}
		uid = append(uid, strconv.Itoa(v))
	}

	return strings.Join(uid, "")
}

func TokenErr2RcErr(tokenErr int) constants.Error {
	switch tokenErr {
	case constants.ERR_INT_OK:
		return constants.RC_OK
	case constants.ERR_INT_TK_DB:
		return constants.RC_SYSTEM_ERR
	case constants.ERR_INT_TK_DUPLICATE:
		return constants.RC_PARAM_ERR
	case constants.ERR_INT_TK_NOTEXISTS:
		return constants.RC_INVALID_TOKEN
	default:
		return constants.RC_SYSTEM_ERR
	}
}
