package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"time"
)

// registerParam holds the request "param" field
type registerParam struct {
	Type    int    `json:"type"`
	Action  string `json:"action"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	VCode   string `json:"vcode"`
	PWD     string `json:"pwd"`
	Spkv    int    `json:"spkv"`
}

// registerRequest holds entire request data
type registerRequest struct {
	Base  common.BaseReq `json:"base"`
	Param registerParam  `json:"param"`
}

// responseData holds response "data" field
type responseRegister struct {
	UID     string `json:"uid"`
	Regtime int64  `json:"regtime"`
}

// registerUserHandler implements the "Echo message" interface
type registerUserHandler struct {
	header       *common.HeaderParams
	registerData *registerRequest
}

func (handler *registerUserHandler) Method() string {
	return http.MethodPost
}

func (handler *registerUserHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK,
			Msg: "ok",
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.registerData)

	// fmt.Println("registerUserHandler) Handle", msg)
	// hashPwd := utils.RsaDecrypt(handler.registerData.Param.PWD, config.GetConfig().PrivKey)

	// TODO: generate a new User ID
	uid := "123456789"

	// newtoken, err := token.New(uid, "key", 24*3600)

	switch handler.registerData.Param.Type {
	case constants.LOGIN_TYPE_UID:
	case constants.LOGIN_TYPE_EMAIL:
	case constants.LOGIN_TYPE_PHONE:
	}

	response.Data = &responseRegister{
		UID:     uid,
		Regtime: time.Now().Unix(),
	}
}
