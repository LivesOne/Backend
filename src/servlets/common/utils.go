package common

import (
	"crypto/rand"
	"encoding/json"
	"hash/crc32"
	"io"
	"net/http"
	"strconv"
	"utils/logger"
)

// FlushJSONData2Client flush json data to http Client
func FlushJSONData2Client(data interface{}, writer http.ResponseWriter) {

	writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	writer.WriteHeader(http.StatusOK)

	if (data == nil) || (writer == nil) {
		return
	}

	// log.Println(" FlushJsonData2Clinet data : ", data)

	toClient, err := json.Marshal(data)
	if err == nil {
		writer.Write(toClient)
		logger.Info("FlushJsonData2Clinet data success:\n", string(toClient))
	} else {
		logger.Info("FlushJsonData2Clinet data error: ", err.Error())
	}
}

// ParseHttpHeaderParams parse the http request header params
func ParseHttpHeaderParams(request *http.Request) *HeaderParams {

	params := &HeaderParams{
		TokenHash: request.Header.Get("Token-Hash"),
		Signature: request.Header.Get("Signature"),
	}

	time := request.Header.Get("Timestamp")
	if len(time) > 0 {
		var err error
		params.Timestamp, err = strconv.ParseInt(time, 10, 64)
		if err != nil {
			// if parse timestamp error, set it as -1
			params.Timestamp = -1
			logger.Info("ParseHttpHeaderParams, parse timestamp error: ", err)
		}
	}

	logger.Info("received http header: ", *params)

	return params
}

// ParseHttpBodyParams parse the http request body params
func ParseHttpBodyParams(request *http.Request, body interface{}) bool {

	logger.Info("request.ContentLength: ", request.ContentLength)
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
			logger.Info("ready body error : ", err)
			return false
		}
	}
	// if (err != io.EOF) || (int64(count) != request.ContentLength) {
	// 	logger.Info("ready body error 9999999999: ", err, count)
	// 	return false
	// }
	logger.Info("ready body: ", bodyparam)

	// bodyparam := request.PostFormValue("param")
	// logger.Info("received http body: ", bodyparam)

	// var data loginRequest
	err := json.Unmarshal([]byte(bodyparam), body)
	if err != nil {
		logger.Info("ParseHttpBodyParams, parse body param error: ", err)
		return false
	}

	return true
}

func GenerateUID(len int) string {
	s := "0123456789"
	box := []byte(s)

	var uid string
	i := 0
	for {
		if i > len-2 {
			break
		}

		r := make([]byte, 16)
		rand.Read(r)
		index := int(r[0]) % 10

		if i == 0 && index == 0 {
			continue
		}

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
