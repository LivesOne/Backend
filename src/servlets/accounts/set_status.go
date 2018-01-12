package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/logger"
)

type setStatusRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *setStatusParam  `json:"param"`
}

type setStatusParam struct {
	Uid    string `json:"uid"`
	Status int    `json:"status"`
}

// checkVCodeHandler
type setStatusHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *checkVCodeRequest   // request body
}

func (handler *setStatusHandler) Method() string {
	return http.MethodPost
}

func (handler *setStatusHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)
	data := setStatusRequest{}
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &data)

	if data.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uidInt64 := utils.Str2Int64(data.Param.Uid)
	err := common.SetAssetStatus(uidInt64, data.Param.Status)

	if err != nil {
		logger.Error("set status error ", err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	}

}
