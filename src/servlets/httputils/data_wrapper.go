/**
 * common request & response data structure
 **/
package httputils

// HeaderParams contains all data in the http request
type HeaderParams struct {
	TokenHash string // header: Token-Hash
	Timestamp int64  // header: Timestamp
	Signature string // header: Signature
	// Data      string // body: params in http body
	// Request *http.Request
	// Writer  http.ResponseWriter
}

// HTTP Request format definitions
// refer to : <LVT_APIs_20171205.docx>
type (
	DeviceReq struct {
		Name string `json:"name,omitempty"`
		DID  string `json:"did,omitempty"`
	}

	AppReq struct {
		Name  string `json:"name,omitempty"`
		AppID string `json:"appid,omitempty"`
		Plat  string `json:"plat,omitempty"`
		Ver   string `json:"ver,omitempty"`
	}

	// BaseReq defines the Request Params format
	// 通用请求格式（Common request format）
	BaseReq struct {
		Device *DeviceReq `json:"device,omitempty"`
		App    *AppReq    `json:"app,omitempty"`
	}

// 	/*
// 		suggest each request handle define this one individually
// 		RequestData struct {
// 			Base  BaseReq     `json:"base,omitempty"`
// 			Param interface{} `json:"param,omitempty"`
// 		}
// 	*/
)

// HTTP Response format definitions
// refer to : <LVT_APIs_20171205.docx>
type (
	BaseResp struct {
		RC  int    `json:"rc"`
		Msg string `json:"msg"`
	}

	// ResponseData defines the http response format
	// 通用返回响应格式（Common response format）
	ResponseData struct {
		Base *BaseResp   `json:"base,omitempty"`
		Data interface{} `json:"data,omitempty"`
	}
)
