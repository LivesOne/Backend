package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)


const(
	MESSAGE = 1
	CALL = 2
	EMAIL = 3
)

type sendVCodeParam struct {
	IMG_id      string `json:"img_id"`
	IMG_vcode   string `json:"img_vcode"`
	Type        int    `json:"type"`
	Action      string `json:"action"`
	Country     int    `json:"country"`
	Phone       string `json:"phone"`
	Check_phone int    `json:"check_phone"`
	EMail       string `json:"email"`
	Ln          string `json:"ln"`
	Expire      int64  `json:"expire"`
}

type sendVCodeRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *sendVCodeParam  `json:"param"`
}

// sendVCodeHandler
type sendVCodeHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *sendVCodeHandler) Method() string {
	return http.MethodPost
}

func (handler *sendVCodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)


	requestData := sendVCodeRequest{}    // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)
	//TODO validate img vcode

	switch requestData.Param.Type {
		case MESSAGE:
			sendMessage(requestData.Param,response)
		case CALL:
			sendCall(requestData.Param,response)
		case EMAIL:
			sendEmail(requestData.Param,response)
	}


}

func sendMessage(param *sendVCodeParam,res *common.ResponseData){
	//TODO send
}

func sendCall(param *sendVCodeParam,res *common.ResponseData){
	//TODO send
}

func sendEmail(param *sendVCodeParam,res *common.ResponseData){
	//TODO send
}