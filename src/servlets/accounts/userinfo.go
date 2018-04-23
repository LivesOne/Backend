package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

type userinfoParam struct {
	Uid string `json:"uid"`
}

type userinfoRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *userinfoParam   `json:"param"`
}

type userinfoResData struct {
	Level    int    `json:"level"`
	NickName string `json:"nick_name"`
}

// sendVCodeHandler
type userinfoHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *userinfoHandler) Method() string {
	return http.MethodPost
}

func (handler *userinfoHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := new(userinfoRequest) // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, requestData)

	base := requestData.Base

	if base == nil || !base.App.IsValid() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	acc,err := common.GetAccountByUID(requestData.Param.Uid)
	if err != nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	response.Data = userinfoResData{
		Level:    acc.Level,
		NickName: acc.Nickname,
	}

}
