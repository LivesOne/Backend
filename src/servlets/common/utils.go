package common

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
		log.Println("FlushJsonData2Clinet data success:\n", string(toClient))
	} else {
		log.Println("FlushJsonData2Clinet data error: ", err.Error())
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
		}
	}

	return params
}
