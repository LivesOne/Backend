package accounts

import (
	"database/sql"
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
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
	Level        int64  `json:"level"`
	NickName     string `json:"nick_name"`
	Country      int64  `json:"country"`
	AvatarUrl    string `json:"avatar_url"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
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

	if param == nil || len(param.Uid) == 0 {
		log.Error("validate param is failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	uid := utils.Str2Int64(param.Uid)
	acc, err := rpc.GetUserInfo(uid)
	if err != nil && err != sql.ErrNoRows {
		log.Error("sql error", err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if acc == nil {
		log.Error("can not find user by uid:", param.Uid)
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	regTime, _ := rpc.GetUserField(uid, microuser.UserField_REGISTER_TIME)
	updTime, _ := rpc.GetUserField(uid, microuser.UserField_UPDATE_TIME)
	response.Data = userinfoResData{
		RegisterTime: utils.GetTs13(utils.Str2Int64(regTime)),
		Level:        acc.Level,
		NickName:     acc.Nickname,
		Country:      acc.Country,
		AvatarUrl:    acc.AvatarUrl,
		Phone:        acc.Phone,
		Email:        acc.Email,
		Ts:           utils.Str2Int64(updTime),
		Hashrate:     common.QueryHashRateByUid(uid),
	}

}
