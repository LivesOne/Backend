package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"database/sql"
	"utils/logger"
)

type userinfoParam struct {
	Uid string `json:"uid"`
}

type userinfoRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *userinfoParam   `json:"param"`
}

type userinfoResData struct {
	Level        int    `json:"level"`
	NickName     string `json:"nick_name"`
	Hashrate     int    `json:"hashrate"`
	RegisterTime int64  `json:"register_time"`
	Ts           int64  `json:"ts"`
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
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
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
		log.Error("validate base is failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	param := requestData.Param

	if param == nil || len(param.Uid) == 0  {
		log.Error("validate param is failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	acc, err := common.GetAccountByUID(param.Uid)
	if err != nil && err != sql.{
		log.Error("sql error",err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}



	response.Data = userinfoResData{
		RegisterTime: utils.GetTs13(acc.RegisterTime),
		Level:        acc.Level,
		NickName:     acc.Nickname,
		Ts:           acc.UpdateTime,
		Hashrate:     common.QueryHashRateByUid(acc.UID),
	}

}
