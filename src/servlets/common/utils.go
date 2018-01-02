package common

import (
	"crypto/rand"
	"encoding/json"
	"hash/crc32"
	"io"
	"net/http"
	"servlets/constants"
	"servlets/token"
	"strconv"
	"time"
	"utils"
	"utils/logger"
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

	var bodyparam string
	var bodyTmp []byte = make([]byte, request.ContentLength)
	// logger.Info("bodyparam: ", len(bodyparam), cap(bodyparam))
	for {
		count, err := request.Body.Read(bodyTmp)
		if count > 0 {
			bodyparam += string(bodyTmp[:count])
		}
		if err == io.EOF {
			break
		}
		// request.Body.Read(bodyparam)
		if err != nil {
			logger.Info("ParseHttpBodyParams: read http body error : ", err)
			return false
		}
	}

	logger.Info("ParseHttpBodyParams: read http body: ", bodyparam)

	err := json.Unmarshal([]byte(bodyparam), body)
	if err != nil {
		logger.Info("ParseHttpBodyParams: parse body param error: ", err)
		return false
	}
	logger.Info("ParseHttpBodyParams: read http request body success:\n", utils.ToJSONIndent(body))

	return true
}

func GenerateUID() string {

	s := "0123456789"
	box := []byte(s)

	len := 9
	uid := "1"

	i := 0
	for {
		if i > len-3 {
			break
		}

		r := make([]byte, 16)
		rand.Read(r)
		index := int(r[0]) % 10

		/*
			if i == 0 && index == 0 {
				continue
			}
		*/

		uid += string(box[index])

		i++
	}

	ieee := crc32.NewIEEE()
	io.WriteString(ieee, uid)
	sum := ieee.Sum32()

	crc := int(sum) % 10

	uid += string(box[crc])

	return uid
}

// GenerateTxID generate a new transaction ID
func GenerateTxID() int64 {
	rid, err := token.GetTxID("id_tx")
	if err != constants.ERR_INT_OK {
		return -1
	}

	const timebase int64 = 1514764800000 // Jan 1, 2018, 00:00:00
	delta := time.Now().UnixNano()/1000/1000 - timebase

	txid := (delta << 22) & 0x7FFFFFFFFFC00000 // move left 22 bit
	txid += int64(rid)

	// fmt.Println("id from redis:", rid, " txid:", txid)

	return txid
}
