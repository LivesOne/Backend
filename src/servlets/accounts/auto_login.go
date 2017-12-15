package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type autologinParam struct {
	token string `json:"token"`
	Key   string `json:"key"`
	Spkv  int    `json:"spkv"`
}

type autologinRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param autologinParam  `json:"param"`
}

type responseAutoLogin struct {
	UID    string `json:"uid"`
	Expire int64  `json:"expire"`
}

// autoLoginHandler implements the "Echo message" interface
type autoLoginHandler struct {
	header    *common.HeaderParams // request header param
	loginData *autologinRequest    // request login data
}

func (handler *autoLoginHandler) Method() string {
	return http.MethodPost
}

func (handler *autoLoginHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK,
			Msg: "ok",
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.loginData)

	// TODO:  get uid && expire data
	uid := "123456789"
	var expire int64 = 24 * 3600

	response.Data = &responseLogin{
		UID:    uid,
		Expire: expire,
	}
}
