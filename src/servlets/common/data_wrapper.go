/**
 * common request & response data structure
 **/
package common

import (
	"servlets/constants"
	"utils/logger"
)

// HeaderParams contains all data in the http request
type HeaderParams struct {
	TokenHash string // header: Token-Hash
	Timestamp int64  // header: Timestamp
	Signature string // header: Signature
}

func (header *HeaderParams) IsValidTimestamp() bool {
	return header.Timestamp > 1
}

func (header *HeaderParams) IsValidTokenhash() bool {
	return len(header.TokenHash) == constants.LEN_HEADER_TOKEN_HASH
}

func (header *HeaderParams) IsValidSign() bool {
	return len(header.Signature) == constants.LEN_HEADER_SIGNATURE
}

func (header *HeaderParams) IsValid() bool {
	return (header.Timestamp > 1) &&
		(len(header.TokenHash) == constants.LEN_HEADER_TOKEN_HASH) &&
		(len(header.Signature) == constants.LEN_HEADER_SIGNATURE)
}

// HTTP Request format definitions
// refer to : <LVT_APIs_20171205.docx>
type (
	DeviceInfo struct {
		Name string `json:"name,omitempty"`
		DID  string `json:"did,omitempty"`
	}

	AppInfo struct {
		Name  string `json:"name,omitempty"`
		AppID string `json:"appid,omitempty"`
		Plat  string `json:"plat,omitempty"`
		Ver   string `json:"ver,omitempty"`
	}

	// BaseReq defines the Request Params format
	// 通用请求格式（Common request format）
	BaseInfo struct {
		Device *DeviceInfo `json:"device,omitempty"`
		App    *AppInfo    `json:"app,omitempty"`
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

// ----------------------------------------------------------------------------

// IsValid check is it a valid App Info
func (app *AppInfo) IsValid() bool {
	return (len(app.Name) > 0) && (len(app.AppID) > 0) && (len(app.Plat) > 0) && (len(app.Ver) > 0)
}

func NewResponseData() *ResponseData {
	return &ResponseData{
		Base: &BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
}

func (responseData *ResponseData) SetResponseBase(error constants.Error) {
	if responseData.Base != nil {
		responseData.Base.RC = error.Rc
		responseData.Base.Msg = error.Msg
	}
	if error != constants.RC_OK {
		logger.Info(error.Msg)
	}
}
