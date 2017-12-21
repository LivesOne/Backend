package common

import (
	"crypto/rand"
	"encoding/json"
	"hash/crc32"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
		Timestamp: -1, // if parse timestamp error, keep it as default: -1
	}

	time := request.Header.Get("Timestamp")
	if len(time) > 0 {
		ts, err := strconv.ParseInt(time, 10, 64)
		if err == nil {
			params.Timestamp = ts
		} else {
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

//发起post请求
func Post(url string, params string) (resBody string, e error) {
	logger.Info("SendPost ---> ", url)
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
